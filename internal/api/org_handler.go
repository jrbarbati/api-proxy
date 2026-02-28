package api

import (
	"api-proxy/internal/model"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (server *Server) handleGetOrgs(w http.ResponseWriter, r *http.Request) {
	active, err := server.orgRepository.FindActive()

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := server.orgRepository.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if route == nil {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	writeJSON(w, route, http.StatusOK)
}

func (server *Server) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	route, err := decodeJSON[model.Org](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := server.orgRepository.Insert(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func (server *Server) handleUpdateOrg(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := decodeJSON[model.Org](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if route.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := server.orgRepository.Update(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
