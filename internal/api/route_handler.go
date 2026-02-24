package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
)

func (server *Server) handleGetRoutes(w http.ResponseWriter, r *http.Request) {
	filter := &repository.RouteFilter{
		Pattern: r.URL.Query().Get("pattern"),
		Method:  r.URL.Query().Get("method"),
	}

	active, err := server.routeRepository.FindActiveByFilter(filter)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	handleGetByID[model.Route](w, r, "route", server.routeRepository.FindByID)
}

func (server *Server) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	handlePost[model.Route](w, r, server.routeRepository.Insert)
}

func (server *Server) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	handlePut[model.Route](w, r, server.routeRepository.Update)
}

func (server *Server) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, server.routeRepository.Delete)
}
