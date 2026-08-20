package httpapi

import (
	"encoding/json"
	"net/http"

	"task133-structload/internal/service"
)

func handleCreateCombination(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		var in service.CreateCombinationInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		c, err := svc.CreateCombination(r.Context(), projectID, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, c)
	}
}

func handleListCombinations(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		cs, err := svc.ListCombinations(r.Context(), projectID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cs)
	}
}

func handleUpdateCombination(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var in service.CreateCombinationInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		c, err := svc.UpdateCombination(r.Context(), id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}
