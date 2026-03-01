package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ServiceAccountHandler struct {
	repository *repository.ServiceAccountRepository
}

func NewServiceAccountHandler(repository *repository.ServiceAccountRepository) *ServiceAccountHandler {
	return &ServiceAccountHandler{repository: repository}
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
	filter := &repository.ServiceAccountFilter{
		Identifier: r.URL.Query().Get("identifier"),
		ClientID:   r.URL.Query().Get("client_id"),
	}

	active, err := sah.repository.FindActiveByFilter(filter)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (sah *ServiceAccountHandler) handleGetServiceAccount(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	sa, err := sah.repository.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if sa == nil {
		http.Error(w, "sa not found", http.StatusNotFound)
		return
	}

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

	created, err := sah.repository.Insert(sa)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

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

	updated, err := sah.repository.Update(sa)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
