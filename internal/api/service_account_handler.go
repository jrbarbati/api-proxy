package api

import (
	"api-proxy/internal/model"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ServiceAccountDataStore interface {
	FindActiveByFilter(filter *model.ServiceAccountFilter) ([]*model.ServiceAccount, error)
	FindByID(id int) (*model.ServiceAccount, error)
	Insert(sa *model.ServiceAccount) (*model.ServiceAccount, error)
	Update(sa *model.ServiceAccount) (*model.ServiceAccount, error)
}

type ServiceAccountHandler struct {
	dataStore ServiceAccountDataStore
}

func NewServiceAccountHandler(serviceAccountDataStore ServiceAccountDataStore) *ServiceAccountHandler {
	return &ServiceAccountHandler{dataStore: serviceAccountDataStore}
}

func (sah *ServiceAccountHandler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", sah.handleGetServiceAccounts)
	r.Get("/{id}", sah.handleGetServiceAccount)
	r.Post("/", sah.handleCreateServiceAccount)
	r.Put("/{id}", sah.handleUpdateServiceAccount)

	return r
}

func (sah *ServiceAccountHandler) handleGetServiceAccounts(w http.ResponseWriter, r *http.Request) {
	filter := &model.ServiceAccountFilter{
		Identifier: r.URL.Query().Get("identifier"),
		ClientID:   r.URL.Query().Get("client_id"),
	}

	active, err := sah.dataStore.FindActiveByFilter(filter)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	for _, sa := range active {
		sa.ClientSecret = ""
	}

	writeJSON(w, active, http.StatusOK)
}

func (sah *ServiceAccountHandler) handleGetServiceAccount(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	sa, err := sah.dataStore.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if sa == nil {
		http.Error(w, "sa not found", http.StatusNotFound)
		return
	}

	sa.ClientSecret = ""

	writeJSON(w, sa, http.StatusOK)
}

func (sah *ServiceAccountHandler) handleCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	sa, err := decodeJSON[model.ServiceAccount](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	hashedSecret, err := hashSecret(sa.ClientSecret)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	sa.ClientSecret = hashedSecret

	created, err := sah.dataStore.Insert(sa)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	created.ClientSecret = ""

	writeJSON(w, created, http.StatusCreated)
}

func (sah *ServiceAccountHandler) handleUpdateServiceAccount(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	sa, err := decodeJSON[model.ServiceAccount](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if sa.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := sah.dataStore.Update(sa)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	updated.ClientSecret = ""

	writeJSON(w, updated, http.StatusOK)
}
