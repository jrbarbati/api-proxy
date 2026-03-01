package api

import (
	"api-proxy/internal/model"
	"api-proxy/internal/repository"
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
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	user, err := iuh.repository.FindByID(uriId)

	if err != nil {
		http.Error(w, "unexpected error.", http.StatusInternalServerError)
		return
	}

	if user == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	user.Password = ""

	writeJSON(w, user, http.StatusOK)
}

func (iuh *InternalUserHandler) handleCreateInternalUser(w http.ResponseWriter, r *http.Request) {
	user, err := decodeJSON[model.InternalUser](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	hashedSecret, err := hashSecret(user.Password)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	user.Password = hashedSecret

	created, err := iuh.repository.Insert(user)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	created.Password = ""

	writeJSON(w, created, http.StatusCreated)
}

func (iuh *InternalUserHandler) handleUpdateInternalUser(w http.ResponseWriter, r *http.Request) {
	uriId, strconvErr := strconv.Atoi(chi.URLParam(r, "id"))

	if strconvErr != nil {
		http.Error(w, "invalid id in the uri", http.StatusBadRequest)
		return
	}

	user, err := decodeJSON[model.InternalUser](r)

	if err != nil {
		http.Error(w, "unable to read json request body", http.StatusBadRequest)
		return
	}

	if user.ID != uriId {
		http.Error(w, "id in uri must match request body id", http.StatusBadRequest)
		return
	}

	updated, err := iuh.repository.Update(user)

	if err != nil {
		http.Error(w, "unexpected error", http.StatusInternalServerError)
		return
	}

	updated.Password = ""

	writeJSON(w, updated, http.StatusOK)
}
