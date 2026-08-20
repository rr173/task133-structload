package httpapi

import (
	"encoding/json"
	"net/http"

	"task133-structload/internal/service"
)

func handleCreateProject(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in service.CreateProjectInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		p, err := svc.CreateProject(r.Context(), in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func handleListProjects(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ps, err := svc.ListProjects(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, ps)
	}
}

func handleGetProject(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		p, err := svc.GetProject(r.Context(), id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}

func handleUpdateProject(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var in service.UpdateProjectInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		p, err := svc.UpdateProject(r.Context(), id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, p)
	}
}

func handleDeleteProject(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := svc.DeleteProject(r.Context(), id); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
	}
}

func handleCreateLevel(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		var in service.CreateLevelInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		l, err := svc.CreateLevel(r.Context(), projectID, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, l)
	}
}

func handleListLevels(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		ls, err := svc.ListLevels(r.Context(), projectID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, ls)
	}
}
