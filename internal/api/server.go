package api

import (
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
	port                     string
	db                       *sql.DB
	routeRepository          *repository.RouteRepository
	orgRepository            *repository.OrgRepository
	serviceAccountRepository *repository.ServiceAccountRepository
	rateLimitRepository      *repository.RateLimitRepository
}

// NewServer creates a server listening on the specified port
func NewServer(
	c *config.Config,
	db *sql.DB,
	routeRepository *repository.RouteRepository,
	orgRepository *repository.OrgRepository,
	serviceAccountRepository *repository.ServiceAccountRepository,
	rateLimitRepository *repository.RateLimitRepository,
) *Server {
	return &Server{
		port:                     c.Server.Port,
		db:                       db,
		routeRepository:          routeRepository,
		orgRepository:            orgRepository,
		serviceAccountRepository: serviceAccountRepository,
		rateLimitRepository:      rateLimitRepository,
	}
}

// Start spins up the server so and registers any handlers as well as provides a graceful shutdown
func (server *Server) Start() error {
	r := chi.NewRouter()

	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Route("/routes", func(r chi.Router) {
			r.Get("/", server.handleGetRoutes)
			r.Get("/{id}", server.handleGetRoute)
			r.Post("/", server.handleCreateRoute)
			r.Put("/{id}", server.handleUpdateRoute)
		})

		r.Route("/orgs", func(r chi.Router) {
			r.Get("/", server.handleGetOrgs)
			r.Get("/{id}", server.handleGetOrg)
			r.Post("/", server.handleCreateOrg)
			r.Put("/{id}", server.handleUpdateOrg)
		})

		r.Route("/serviceAccounts", func(r chi.Router) {
			r.Get("/", server.handleGetServiceAccounts)
			r.Get("/{id}", server.handleGetServiceAccount)
			r.Post("/", server.handleCreateServiceAccount)
			r.Put("/{id}", server.handleUpdateServiceAccount)
		})

		r.Route("/rateLimits", func(r chi.Router) {
			r.Get("/", server.handleGetRateLimits)
			r.Get("/{id}", server.handleGetRateLimit)
			r.Post("/", server.handleCreateRateLimit)
			r.Put("/{id}", server.handleUpdateRateLimit)
		})
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
