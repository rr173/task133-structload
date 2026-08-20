package httpapi

import (
	"net/http"
	"testing"
)

func TestBug17_MissingProjectReviewIsNotAnEmptyReport(t *testing.T) {
	_, h := newAPI(t)
	code, _ := doJSON(t, h, http.MethodPost, "/api/projects", map[string]any{"code": "P-17", "name": "existing", "exposure": "C", "importance": "II", "wind_speed": 380})
	if code != http.StatusCreated { t.Fatalf("create status = %d", code) }
	code, _ = doJSON(t, h, http.MethodGet, "/api/projects/does-not-exist/review", nil)
	if code != http.StatusNotFound { t.Fatalf("missing project review status = %d, want 404", code) }
}
