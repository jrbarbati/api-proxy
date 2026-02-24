package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
)

func (server *Server) handleGetServiceAccounts(w http.ResponseWriter, r *http.Request) {
	filter := &repository.ServiceAccountFilter{
		Identifier: r.URL.Query().Get("identifier"),
		ClientID:   r.URL.Query().Get("clientId"),
	}

	active, err := server.serviceAccountRepository.FindActiveByFilter(filter)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetServiceAccount(w http.ResponseWriter, r *http.Request) {
	handleGetByID[model.ServiceAccount](w, r, "service account", server.serviceAccountRepository.FindByID)
}

func (server *Server) handleCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	handlePost[model.ServiceAccount](w, r, server.serviceAccountRepository.Insert)
}

func (server *Server) handleUpdateServiceAccount(w http.ResponseWriter, r *http.Request) {
	handlePut[model.ServiceAccount](w, r, server.serviceAccountRepository.Update)
}

func (server *Server) handleDeleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, server.serviceAccountRepository.Delete)
}
