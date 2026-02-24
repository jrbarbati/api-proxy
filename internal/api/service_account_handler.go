package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (server *Server) handleGetServiceAccounts(w http.ResponseWriter, r *http.Request) {
	filter := &repository.ServiceAccountFilter{
		Identifier: r.URL.Query().Get("identifier"),
		ClientID:   r.URL.Query().Get("client_id"),
	}

	active, err := server.serviceAccountRepository.FindActiveByFilter(filter)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetServiceAccount(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		writeError(w, newError("invalid id in the uri", http.StatusBadRequest))
		return
	}

	route, err := server.serviceAccountRepository.FindByID(uriId)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	if route == nil {
		writeError(w, newError("route not found", http.StatusNotFound))
		return
	}

	writeJSON(w, route, http.StatusOK)
}

func (server *Server) handleCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	route, err := decodeJSON[model.ServiceAccount](r)

	if err != nil {
		writeError(w, newError("unable to read json request body", http.StatusBadRequest))
		return
	}

	created, err := server.serviceAccountRepository.Insert(route)

	if err != nil {
		writeError(w, newError("unexpected error", http.StatusInternalServerError))
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func (server *Server) handleUpdateServiceAccount(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		writeError(w, newError("invalid id in the uri", http.StatusBadRequest))
		return
	}

	route, err := decodeJSON[model.ServiceAccount](r)

	if err != nil {
		writeError(w, newError("unable to read json request body", http.StatusBadRequest))
		return
	}

	if route.ID != uriId {
		writeError(w, newError("id in uri must match request body id", http.StatusBadRequest))
		return
	}

	updated, err := server.serviceAccountRepository.Update(route)

	if err != nil {
		writeError(w, newError("unexpected error", http.StatusInternalServerError))
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
