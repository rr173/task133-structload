package httpapi

import (
	"encoding/json"
	"net/http"

	"task133-structload/internal/service"
)

func handleAddLoadCase(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		componentID := r.PathValue("id")
		var in service.AddManualLoadCaseInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		lc, err := svc.AddManualLoadCase(r.Context(), componentID, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, lc)
	}
}

func handleListLoadCases(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		componentID := r.PathValue("id")
		lcs, err := svc.ListLoadCases(r.Context(), componentID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, lcs)
	}
}

func handleListDerivedCases(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		componentID := r.PathValue("id")
		lcs, err := svc.ListDerivedCases(r.Context(), componentID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, lcs)
	}
}
