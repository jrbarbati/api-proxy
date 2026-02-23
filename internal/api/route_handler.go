package api

import (
	"api-proxy/internal/model"
	"encoding/json"
	"net/http"
)

func (server *Server) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var route model.Route

	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	created, err := server.routeRepository.Insert(&route)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(created)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
