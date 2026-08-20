package httpapi

import (
	"net/http"

	"task133-structload/internal/service"
)

// handleReview exposes the deterministic design-review report. It is kept
// separate from /summary because a review can identify missing evidence even
// when an aggregate summary looks benign.
func handleReview(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		report, err := svc.Review(r.Context(), projectID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, report)
	}
}
