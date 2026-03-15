package api

import (
	"api-proxy/internal/model"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type RequestDataStorer interface {
	FindBetween(start time.Time, end time.Time) ([]*model.Request, error)
}

type RequestHandler struct {
	datastore RequestDataStorer
}

func NewRequestHandler(datastore RequestDataStorer) *RequestHandler {
	return &RequestHandler{
		datastore: datastore,
	}
}

func (rh *RequestHandler) Router() http.Handler {
	router := chi.NewRouter()

	router.Get("/", rh.HandleGetRequests)

	return router
}

func (rh *RequestHandler) HandleGetRequests(w http.ResponseWriter, r *http.Request) {
	from, fromErr := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	to, toErr := time.Parse(time.RFC3339, r.URL.Query().Get("to"))

	if fromErr != nil || toErr != nil {
		slog.Error("unable to parse url param(s)", "from", from, "to", to)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if to.Before(from) {
		slog.Error("invalid ordering of params", "from", from, "to", to)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	requests, err := rh.datastore.FindBetween(from, to)

	if err != nil {
		slog.Error("unable to find requests", "from", from, "to", to)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeJSON(w, requests, http.StatusOK)
}
