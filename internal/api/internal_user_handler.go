package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type InternalUserHandler struct {
	repository *repository.InternalUserRepository
}

func NewInternalUserHandler(repository *repository.InternalUserRepository) *InternalUserHandler {
	return &InternalUserHandler{repository: repository}
}

func (iuh *InternalUserHandler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", iuh.handleGetInternalUsers)
	r.Get("/{id}", iuh.handleGetInternalUser)
	r.Post("/", iuh.handleCreateInternalUser)
	r.Put("/{id}", iuh.handleUpdateInternalUser)

	return r
}

func (iuh *InternalUserHandler) handleGetInternalUsers(w http.ResponseWriter, r *http.Request) {
	filter := &repository.InternalUserFilter{
		Email: r.URL.Query().Get("email"),
	}

	active, err := iuh.repository.FindActive(filter)

	if err != nil {
		slog.Error("error finding active internal users", "error", err)
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	for _, user := range active {
		user.Password = ""
	}

	writeJSON(w, active, http.StatusOK)
}

func (iuh *InternalUserHandler) handleGetInternalUser(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		slog.Error("error converting id to int while GETing internal user", "id", chi.URLParam(r, "id"), "error", strconvErr)
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	user, err := iuh.repository.FindByID(uriId)

	if err != nil {
		slog.Error("error finding internal user", "id", uriId, "error", err)
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if user == nil {
		slog.Error("internal user not found with id", "id", uriId)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	user.Password = ""

	writeJSON(w, user, http.StatusOK)
}

func (iuh *InternalUserHandler) handleCreateInternalUser(w http.ResponseWriter, r *http.Request) {
	user, err := decodeJSON[model.InternalUser](r)

	if err != nil {
		slog.Error("error decoding json while creating user", "error", err)
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	hashedSecret, err := hashSecret(user.Password)

	if err != nil {
		slog.Error("error hashing password for internal user", "error", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	user.Password = hashedSecret

	created, err := iuh.repository.Insert(user)

	if err != nil {
		slog.Error("error inserting internal user", "error", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	created.Password = ""

	writeJSON(w, created, http.StatusCreated)
}

func (iuh *InternalUserHandler) handleUpdateInternalUser(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		slog.Error("error converting id to int while PUTing internal user", "id", chi.URLParam(r, "id"), "error", strconvErr)
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	user, err := decodeJSON[model.InternalUser](r)

	if err != nil {
		slog.Error("error decoding json while updating user", "error", err)
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if user.ID != uriId {
		slog.Error("internal user not found for update with id", "id", uriId, "user", user)
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := iuh.repository.Update(user)

	if err != nil {
		slog.Error("error updating internal user", "error", err)
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	updated.Password = ""

	writeJSON(w, updated, http.StatusOK)
}
