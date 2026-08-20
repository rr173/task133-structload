package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"task133-structload/internal/model"
	"task133-structload/internal/service"
)

// errInvalid wraps a message in model.ErrInvalid for 422 responses.
func errInvalid(msg string) error {
	return errors.New(msg + ": " + model.ErrInvalid.Error())
}

// invalid wraps a message as a 422 invalid error (kept for brevity).
func invalid(msg string) error { return errInvalid(msg) }

func handleCreateComponent(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		levelID := r.PathValue("id")
		var in service.CreateComponentInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		in.LevelID = levelID
		c, err := svc.CreateComponent(r.Context(), "", in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, c)
	}
}

func handleListComponentsByLevel(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		levelID := r.PathValue("id")
		cs, err := svc.ListComponentsByLevel(r.Context(), levelID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cs)
	}
}

func handleListComponentsByProject(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("id")
		cs, err := svc.ListComponentsByProject(r.Context(), projectID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, cs)
	}
}

func handleUpdateComponent(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var in service.UpdateComponentInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, errInvalid("decode body"))
			return
		}
		c, err := svc.UpdateComponent(r.Context(), id, in)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}

func handleGetComponent(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		c, err := svc.GetComponent(r.Context(), id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}
