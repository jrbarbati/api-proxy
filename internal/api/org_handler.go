package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type OrgHandler struct {
	repository *repository.OrgRepository
}

func NewOrgHandler(repository *repository.OrgRepository) *OrgHandler {
	return &OrgHandler{repository: repository}
}

func (oh *OrgHandler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", oh.handleGetOrgs)
	r.Get("/{id}", oh.handleGetOrg)
	r.Post("/", oh.handleCreateOrg)
	r.Put("/{id}", oh.handleUpdateOrg)

	return r
}

func (oh *OrgHandler) handleGetOrgs(w http.ResponseWriter, r *http.Request) {
	active, err := oh.repository.FindActive()

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (oh *OrgHandler) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := oh.repository.FindByID(uriId)

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

func (oh *OrgHandler) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	route, err := decodeJSON[model.Org](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := oh.repository.Insert(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func (oh *OrgHandler) handleUpdateOrg(w http.ResponseWriter, r *http.Request) {
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

	updated, err := oh.repository.Update(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
