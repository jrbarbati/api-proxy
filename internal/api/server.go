package api

import (
	"api-proxy/internal/config"
	"api-proxy/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// Server represents an HTTP server with graceful shutdown support.
type Server struct {
	port            string
	db              *sql.DB
	routeRepository *repository.RouteRepository
}

// NewServer creates a server listening on the specified port
func NewServer(c *config.Config, db *sql.DB, routeRepository *repository.RouteRepository) *Server {
	return &Server{
		port:            c.Server.Port,
		db:              db,
		routeRepository: routeRepository,
	}
}

// Start spins up the server so and registers any handlers as well as provides a graceful shutdown
func (server *Server) Start() error {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/routes", func(r chi.Router) {
			r.Post("/", server.handleCreateRoute)
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
