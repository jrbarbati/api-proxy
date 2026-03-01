package api

import (
	"api-proxy/internal/api/middleware"
	"api-proxy/internal/config"
	"api-proxy/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	}
}

// Start spins up the server so and registers any handlers as well as provides a graceful shutdown
func (server *Server) Start() error {
	r := chi.NewRouter()

	sar := repository.NewServiceAccountRepository(server.db)
	iur := repository.NewInternalUserRepository(server.db)

	authHandler := NewAuthHandler(server.jwtSigningSecret, server.adminJwtSigningSecret, sar, iur)
	internalUserHandler := NewInternalUserHandler(iur)
	orgHandler := NewOrgHandler(repository.NewOrgRepository(server.db))
	rateLimitHandler := NewRateLimitHandler(repository.NewRateLimitRepository(server.db))
	routeHandler := NewRouteHandler(repository.NewRouteRepository(server.db))
	serviceAccountHandler := NewServiceAccountHandler(sar)

	r.Post("/api/v1/oauth/token", authHandler.handleOAuth)
	r.Post("/api/v1/admin/oauth/token", authHandler.handleInternalOAuth)

	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Use(middleware.AdminAuth(server.adminJwtSigningSecret))

		r.Mount("/users", internalUserHandler.Router())
		r.Mount("/orgs", orgHandler.Router())
		r.Mount("/rate-limits", rateLimitHandler.Router())
		r.Mount("/routes", routeHandler.Router())
		r.Mount("/service-accounts", serviceAccountHandler.Router())
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpServer := server.listenAndServe(r)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return httpServer.Shutdown(shutdownCtx)
}

func (server *Server) listenAndServe(r *chi.Mux) *http.Server {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%v", server.port), Handler: r}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %s\n", err)
		}
	}()

	return httpServer
}

// Port returns the port the server is listening on
func (server *Server) Port() string {
	return server.port
}
