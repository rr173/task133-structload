package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"task133-structload/internal/model"
	"task133-structload/internal/service"
	"task133-structload/internal/webfs"
)

// New builds the HTTP mux wiring all routes to the service.
func New(svc *service.Service) http.Handler {
	mux := http.NewServeMux()

	// Frontend (embedded).
	mux.Handle("GET /", webfs.Handler())

	// Projects.
	mux.HandleFunc("POST /api/projects", handleCreateProject(svc))
	mux.HandleFunc("GET /api/projects", handleListProjects(svc))
	mux.HandleFunc("GET /api/projects/{id}", handleGetProject(svc))
	mux.HandleFunc("PATCH /api/projects/{id}", handleUpdateProject(svc))
	mux.HandleFunc("DELETE /api/projects/{id}", handleDeleteProject(svc))

	// Levels.
	mux.HandleFunc("POST /api/projects/{id}/levels", handleCreateLevel(svc))
	mux.HandleFunc("GET /api/projects/{id}/levels", handleListLevels(svc))

	// Components.
	mux.HandleFunc("POST /api/levels/{id}/components", handleCreateComponent(svc))
	mux.HandleFunc("GET /api/levels/{id}/components", handleListComponentsByLevel(svc))
	mux.HandleFunc("GET /api/projects/{id}/components", handleListComponentsByProject(svc))
	mux.HandleFunc("PATCH /api/components/{id}", handleUpdateComponent(svc))
	mux.HandleFunc("GET /api/components/{id}", handleGetComponent(svc))

	// Load cases.
	mux.HandleFunc("POST /api/components/{id}/loadcases", handleAddLoadCase(svc))
	mux.HandleFunc("GET /api/components/{id}/loadcases", handleListLoadCases(svc))
	mux.HandleFunc("GET /api/components/{id}/derived", handleListDerivedCases(svc))

	// Combinations.
	mux.HandleFunc("POST /api/projects/{id}/combinations", handleCreateCombination(svc))
	mux.HandleFunc("GET /api/projects/{id}/combinations", handleListCombinations(svc))
	mux.HandleFunc("PATCH /api/combinations/{id}", handleUpdateCombination(svc))

	// Checks.
	mux.HandleFunc("POST /api/projects/{id}/checks/run", handleRunChecks(svc))
	mux.HandleFunc("GET /api/projects/{id}/checks", handleListChecksByProject(svc))
	mux.HandleFunc("GET /api/components/{id}/checks", handleListChecksByComponent(svc))
	mux.HandleFunc("GET /api/projects/{id}/summary", handleSummary(svc))
	mux.HandleFunc("GET /api/projects/{id}/review", handleReview(svc))

	// Derive / recompute / audit / override / event log / replay.
	mux.HandleFunc("POST /api/projects/{id}/derive", handleDerive(svc))
	mux.HandleFunc("POST /api/projects/{id}/recompute", handleReconcile(svc))
	mux.HandleFunc("GET /api/projects/{id}/audit", handleAudit(svc))
	mux.HandleFunc("POST /api/components/{id}/override", handleOverride(svc))
	mux.HandleFunc("GET /api/projects/{id}/overrides", handleListOverrides(svc))
	mux.HandleFunc("GET /api/projects/{id}/eventlog", handleListEvents(svc))
	mux.HandleFunc("POST /api/admin/replay", handleReplay(svc))

	return logging(mux)
}

// logging wraps the mux with a minimal access log to stderr.
func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// no per-request logging noise for smoke-test determinism
		}
		h.ServeHTTP(w, r)
	})
}

// writeJSON encodes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps a service error to an HTTP response.
func writeError(w http.ResponseWriter, err error) {
	code := model.HTTPCode(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "code": http.StatusText(code)})
}
