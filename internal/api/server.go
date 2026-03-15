package api

import (
	"api-proxy/internal/api/middleware"
	"api-proxy/internal/cache"
	"api-proxy/internal/config"
	"api-proxy/internal/logger"
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

// Server represents an HTTP server with graceful shutdown support.
type Server struct {
	jwtSigningSecret      string
	adminJwtSigningSecret string
	port                  string
	db                    *sql.DB
	requestLogQueueSize   int
}

// NewServer creates a server listening on the specified port
func NewServer(
	c *config.Config,
	db *sql.DB,
) *Server {
	return &Server{
		port:                  c.Server.Port,
		jwtSigningSecret:      c.JWTConfig.SigningSecret,
		adminJwtSigningSecret: c.JWTConfig.Admin.SigningSecret,
		db:                    db,
		requestLogQueueSize:   *c.LoggingConfig.LoggingRequestConfig.QueueSize,
	}
}

// Start spins up the server so and registers any handlers as well as provides a graceful shutdown
func (server *Server) Start() error {
	router := chi.NewRouter()

	internalUserRepo := repository.NewInternalUserRepository(server.db)
	orgRepo := repository.NewOrgRepository(server.db)
	rateLimitRepo := repository.NewRateLimitRepository(server.db)
	routeRepo := repository.NewRouteRepository(server.db)
	routeCache := cache.NewRouteCache()
	rateLimitCache := cache.NewRateLimitCache()
	serviceAccountRepo := repository.NewServiceAccountRepository(server.db)
	requestRepo := repository.NewRequestRepository(server.db)

	requestLogger := logger.NewRequestLogger(requestRepo, server.requestLogQueueSize)

	authHandler := NewAuthHandler(server.jwtSigningSecret, server.adminJwtSigningSecret, serviceAccountRepo, internalUserRepo)

	router.Post("/api/v1/oauth/token", authHandler.handleOAuth)
	router.Post("/api/v1/admin/oauth/token", authHandler.handleInternalOAuth)

	router.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(middleware.AdminAuth(server.adminJwtSigningSecret))

		r.Mount("/users", NewInternalUserHandler(internalUserRepo).Router())
		r.Mount("/orgs", NewOrgHandler(orgRepo).Router())
		r.Mount("/rate-limits", NewRateLimitHandler(rateLimitRepo).Router())
		r.Mount("/routes", NewRouteHandler(routeRepo).Router())
		r.Mount("/service-accounts", NewServiceAccountHandler(serviceAccountRepo).Router())
	})

	router.With(
		middleware.LogRequest(requestLogger),
		middleware.ExternalAuth(server.jwtSigningSecret),
		middleware.RateLimit(rateLimitCache),
		middleware.ResolveRoute(routeCache),
	).Handle("/*", NewProxyHandler())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	requestLogger.Start(ctx)
	routeCache.StartSync(ctx, 2*time.Minute, func() ([]*model.Route, error) { // TODO: Do some benchmarking on routeRepo.FindActiveByFilter and/orgRepo the syncCache() method and adjust the interval accordingly
		return routeRepo.FindActiveByFilter(nil)
	})
	rateLimitCache.StartSync(ctx, 5*time.Minute, func() ([]*model.RateLimit, error) {
		return rateLimitRepo.FindActiveByFilter(nil)
	})
	httpServer := server.listenAndServe(router)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return httpServer.Shutdown(shutdownCtx)
}

func (server *Server) listenAndServe(r *chi.Mux) *http.Server {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%v", server.port), Handler: r}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	return httpServer
}

// Port returns the port the server is listening on
func (server *Server) Port() string {
	return server.port
}
