package httpapi

import (
	"net/http"
	"testing"
)

func TestBug18_ProjectSnowRevisionIsImmediatelyAndPersistentlyVisible(t *testing.T) {
	_, h := newAPI(t)
	code, body := doJSON(t, h, http.MethodPost, "/api/projects", map[string]any{"code": "P-18", "name": "snow update", "exposure": "C", "importance": "II", "wind_speed": 380, "ground_snow": 20})
	if code != http.StatusCreated { t.Fatalf("create status = %d", code) }
	id := body["id"].(string)
	code, body = doJSON(t, h, http.MethodPatch, "/api/projects/"+id, map[string]any{"ground_snow": 95})
	if code != http.StatusOK || body["ground_snow"] != float64(95) { t.Fatalf("update status/body = %d %#v", code, body) }
	code, body = doJSON(t, h, http.MethodGet, "/api/projects/"+id, nil)
	if code != http.StatusOK || body["ground_snow"] != float64(95) { t.Fatalf("read status/body = %d %#v", code, body) }
}
