package api

import (
	"api-proxy/internal/model"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func handleGetByID[T model.Identifiable](w http.ResponseWriter, r *http.Request, requestType string, findById func(id int) (*T, error)) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		writeError(w, newError("invalid id in the uri", http.StatusBadRequest))
		return
	}

	result, err := findById(uriId)

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	if result == nil {
		writeError(w, newError(requestType+" not found", http.StatusNotFound))
		return
	}

	writeJSON(w, result, http.StatusOK)
}

func handlePut[T model.Identifiable](w http.ResponseWriter, r *http.Request, update func(req *T) (*T, error)) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		writeError(w, newError("invalid id in the uri", http.StatusBadRequest))
		return
	}

	request, err := decodeJSON[T](r)

	if err != nil {
		writeError(w, newError("unable to read json request body", http.StatusBadRequest))
		return
	}

	if request.GetID() != uriId {
		writeError(w, newError("id in uri must match request body id", http.StatusBadRequest))
		return
	}

	updated, err := update(&request)

	if err != nil {
		writeError(w, newError("unexpected error", http.StatusInternalServerError))
		return
	}

	writeJSON(w, updated, http.StatusOK)
}

func handlePost[T model.Identifiable](w http.ResponseWriter, r *http.Request, insert func(req *T) (*T, error)) {
	request, err := decodeJSON[T](r)

	if err != nil {
		writeError(w, newError("unable to read json request body", http.StatusBadRequest))
		return
	}

	created, err := insert(&request)

	if err != nil {
		writeError(w, newError("unexpected error", http.StatusInternalServerError))
		return
	}

	writeJSON(w, created, http.StatusCreated)
}

func handleDelete(w http.ResponseWriter, r *http.Request, delete func(id int) error) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		writeError(w, newError("invalid id in the uri", http.StatusBadRequest))
		return
	}

	err := delete(uriId)

	if err != nil {
		writeError(w, newError("unexpected error", http.StatusInternalServerError))
		return
	}

	emptyResponse(w, http.StatusNoContent)
}

func decodeJSON[T any](r *http.Request) (T, error) {
	var body T

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return body, err
	}

	return body, nil
}
