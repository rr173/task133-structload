package httpapi

import (
	"encoding/json"
	"net/http"

	"task133-structload/internal/service"
)

func handleOverride(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		componentID := r.PathValue("id")
		var in service.OverrideComponentInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		o, err := svc.OverrideComponent(r.Context(), componentID, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, o)
	}
}

func handleListOverrides(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		os, err := svc.ListOverrides(r.Context(), projectID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, os)
	}
}
