package api

import (
	"api-proxy/internal/model"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RateLimitDataStore interface {
	FindActiveByFilter(filter *model.RateLimitFilter) ([]*model.RateLimit, error)
	FindByID(id int) (*model.RateLimit, error)
	Insert(rateLimit *model.RateLimit) (*model.RateLimit, error)
	Update(rateLimit *model.RateLimit) (*model.RateLimit, error)
}

type RateLimitHandler struct {
	dataStore RateLimitDataStore
}

func NewRateLimitHandler(rateLimitDataStore RateLimitDataStore) *RateLimitHandler {
	return &RateLimitHandler{dataStore: rateLimitDataStore}
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
	orgID, orgIdParamErr := queryParam("orgId", r, toIntParam)
	serviceAccountID, saIDParamErr := queryParam("serviceAccountId", r, toIntParam)

	if orgIdParamErr != nil || saIDParamErr != nil {
		slog.Error("either orgId or serviceAccountId was invalid", "org_id", orgIdParamErr, "service_account_id", saIDParamErr)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	filter := &model.RateLimitFilter{
		OrgId:            orgID,
		ServiceAccountId: serviceAccountID,
	}

	active, err := rlh.dataStore.FindActiveByFilter(filter)

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

	rateLimit, err := rlh.dataStore.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if rateLimit == nil {
		http.Error(w, "rate limit not found", http.StatusNotFound)
		return
	}

	writeJSON(w, rateLimit, http.StatusOK)
}

func (rlh *RateLimitHandler) handleCreateRateLimit(w http.ResponseWriter, r *http.Request) {
	rateLimit, err := decodeJSON[model.RateLimit](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := rlh.dataStore.Insert(rateLimit)

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

	rateLimit, err := decodeJSON[model.RateLimit](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if rateLimit.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := rlh.dataStore.Update(rateLimit)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
