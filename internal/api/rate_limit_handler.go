package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RateLimitHandler struct {
	repository *repository.RateLimitRepository
}

func NewRateLimitHandler(repository *repository.RateLimitRepository) *RateLimitHandler {
	return &RateLimitHandler{repository: repository}
}

func (rlh *RateLimitHandler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", rlh.handleGetRateLimits)
	r.Get("/{id}", rlh.handleGetRateLimit)
	r.Post("/", rlh.handleCreateRateLimit)
	r.Put("/{id}", rlh.handleUpdateRateLimit)

	return r
}

func (rlh *RateLimitHandler) handleGetRateLimits(w http.ResponseWriter, r *http.Request) {
	filter := &repository.RateLimitFilter{
		OrgId:            r.URL.Query().Get("orgId"),
		ServiceAccountId: r.URL.Query().Get("serviceAccountId"),
	}

	active, err := rlh.repository.FindActiveByFilter(filter)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (rlh *RateLimitHandler) handleGetRateLimit(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := rlh.repository.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if route == nil {
		http.Error(w, "rate limit not found", http.StatusNotFound)
		return
	}

	writeJSON(w, route, http.StatusOK)
}

func (rlh *RateLimitHandler) handleCreateRateLimit(w http.ResponseWriter, r *http.Request) {
	route, err := decodeJSON[model.RateLimit](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := rlh.repository.Insert(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func (rlh *RateLimitHandler) handleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := decodeJSON[model.RateLimit](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if route.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := rlh.repository.Update(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
