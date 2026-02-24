package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
)

func (server *Server) handleGetRateLimits(w http.ResponseWriter, r *http.Request) {
	filter := &repository.RateLimitFilter{
		OrgId:            r.URL.Query().Get("orgId"),
		ServiceAccountId: r.URL.Query().Get("serviceAccountId"),
	}

	active, err := server.rateLimitRepository.FindActiveByFilter(filter)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetRateLimit(w http.ResponseWriter, r *http.Request) {
	handleGetByID[model.RateLimit](w, r, "service account", server.rateLimitRepository.FindByID)
}

func (server *Server) handleCreateRateLimit(w http.ResponseWriter, r *http.Request) {
	handlePost[model.RateLimit](w, r, server.rateLimitRepository.Insert)
}

func (server *Server) handleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	handlePut[model.RateLimit](w, r, server.rateLimitRepository.Update)
}

func (server *Server) handleDeleteRateLimit(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, server.rateLimitRepository.Delete)
}
