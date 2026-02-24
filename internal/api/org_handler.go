package api

import (
	"api-proxy/internal/model"
	"net/http"
)

func (server *Server) handleGetOrgs(w http.ResponseWriter, r *http.Request) {
	active, err := server.orgRepository.FindActive()

	if err != nil {
		writeError(w, newError("unexpected error.", http.StatusInternalServerError))
		return
	}

	writeJSON(w, active, http.StatusOK)
}

func (server *Server) handleGetOrg(w http.ResponseWriter, r *http.Request) {
	handleGetByID[model.Org](w, r, "org", server.orgRepository.FindByID)
}

func (server *Server) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	handlePost[model.Org](w, r, server.orgRepository.Insert)
}

func (server *Server) handleUpdateOrg(w http.ResponseWriter, r *http.Request) {
	handlePut[model.Org](w, r, server.orgRepository.Update)
}

func (server *Server) handleDeleteOrg(w http.ResponseWriter, r *http.Request) {
	handleDelete(w, r, server.orgRepository.Delete)
}
