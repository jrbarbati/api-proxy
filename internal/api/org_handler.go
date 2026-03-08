package api

import (
	"api-proxy/internal/model"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type OrgDataStorer interface {
	FindActive() ([]*model.Org, error)
	FindByID(id int) (*model.Org, error)
	Insert(org *model.Org) (*model.Org, error)
	Update(org *model.Org) (*model.Org, error)
}

type OrgHandler struct {
	dataStore OrgDataStorer
}

func NewOrgHandler(orgDataStore OrgDataStorer) *OrgHandler {
	return &OrgHandler{dataStore: orgDataStore}
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
	active, err := oh.dataStore.FindActive()

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

	org, err := oh.dataStore.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if org == nil {
		http.Error(w, "org not found", http.StatusNotFound)
		return
	}

	writeJSON(w, org, http.StatusOK)
}

func (oh *OrgHandler) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	org, err := decodeJSON[model.Org](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := oh.dataStore.Insert(org)

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

	org, err := decodeJSON[model.Org](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if org.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := oh.dataStore.Update(org)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
