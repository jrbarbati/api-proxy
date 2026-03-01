package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RouteHandler struct {
	repository *repository.RouteRepository
}

func NewRouteHandler(repository *repository.RouteRepository) *RouteHandler {
	return &RouteHandler{repository: repository}
}

func (rh *RouteHandler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", rh.handleGetRoutes)
	r.Get("/{id}", rh.handleGetRoute)
	r.Post("/", rh.handleCreateRoute)
	r.Put("/{id}", rh.handleUpdateRoute)

	return r
}

func (rh *RouteHandler) handleGetRoutes(w http.ResponseWriter, r *http.Request) {
	filter := &repository.RouteFilter{
		Pattern: r.URL.Query().Get("pattern"),
		Method:  r.URL.Query().Get("method"),
	}

	active, err := rh.repository.FindActiveByFilter(filter)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (rh *RouteHandler) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := rh.repository.FindByID(uriId)

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

func (rh *RouteHandler) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	route, err := decodeJSON[model.Route](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	created, err := rh.repository.Insert(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func (rh *RouteHandler) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	route, err := decodeJSON[model.Route](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if route.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := rh.repository.Update(route)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, updated, http.StatusOK)
}
