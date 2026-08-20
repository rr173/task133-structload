package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/service"
	"task133-structload/internal/store"
)

func newAPI(t *testing.T) (*service.Service, http.Handler) {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/api.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	svc := service.New(s, clock.NewFake(time.UnixMilli(1740000000000)))
	return svc, New(svc)
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	var req *http.Request
	if rdr != nil {
		req = httptest.NewRequest(method, path, rdr)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	var out map[string]any
	if w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), &out)
	}
	return w.Code, out
}

func TestCreateProjectHTTP(t *testing.T) {
	_, h := newAPI(t)
	code, body := doJSON(t, h, "POST", "/api/projects", map[string]any{
		"code": "X", "name": "X", "exposure": "C", "importance": "II", "wind_speed": 400, "ground_snow": 95,
	})
	if code != 201 {
		t.Fatalf("expected 201, got %d: %v", code, body)
	}
	if body["id"] == nil || body["code"] != "X" {
		t.Errorf("unexpected project body: %v", body)
	}
}

func TestCreateProjectValidationHTTP(t *testing.T) {
	_, h := newAPI(t)
	code, body := doJSON(t, h, "POST", "/api/projects", map[string]any{
		"code": "X", "name": "X", "exposure": "Z", "importance": "II", "wind_speed": 400, "ground_snow": 0,
	})
	if code != 422 {
		t.Errorf("expected 422 for invalid exposure, got %d: %v", code, body)
	}
}

func TestNotFoundProject(t *testing.T) {
	_, h := newAPI(t)
	code, _ := doJSON(t, h, "GET", "/api/projects/nonexistent", nil)
	if code != 404 {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestFullFlowHTTP(t *testing.T) {
	svc, h := newAPI(t)
	code, body := doJSON(t, h, "POST", "/api/projects", map[string]any{
		"code": "FP", "name": "FP", "exposure": "C", "importance": "II", "wind_speed": 400, "ground_snow": 95,
	})
	pid := body["id"].(string)
	if code != 201 || pid == "" {
		t.Fatal("create project failed")
	}
	// level
	_, lvl := doJSON(t, h, "POST", "/api/projects/"+pid+"/levels", map[string]any{"name": "1F", "elevation": 10000, "gross_area": 1000000})
	lvlID := lvl["id"].(string)
	// component
	_, cmp := doJSON(t, h, "POST", "/api/levels/"+lvlID+"/components", map[string]any{
		"code": "B1", "type": "beam", "span": 6000, "tributary_area": 400000, "tributary_width": 2500,
		"kll": 2, "nominal_moment": 100000, "nominal_shear": 100000, "phi_b": 90, "phi_c": 90, "phi_v": 90,
		"wind_surface": "windward", "roof_slope": 0,
	})
	cid := cmp["id"].(string)
	// load cases
	doJSON(t, h, "POST", "/api/components/"+cid+"/loadcases", map[string]any{"kind": "D", "load_type": "area_pressure", "magnitude": 150, "direction": "gravity"})
	doJSON(t, h, "POST", "/api/components/"+cid+"/loadcases", map[string]any{"kind": "L", "load_type": "area_pressure", "magnitude": 240, "direction": "gravity"})
	// combination
	doJSON(t, h, "POST", "/api/projects/"+pid+"/combinations", map[string]any{"name": "c", "kind": "strength", "coeff_d": 120, "coeff_l": 160, "coeff_s": 50})
	// run checks
	code, _ = doJSON(t, h, "POST", "/api/projects/"+pid+"/checks/run", nil)
	if code != 200 {
		t.Fatal("run checks failed")
	}
	// summary
	code, sum := doJSON(t, h, "GET", "/api/projects/"+pid+"/summary", nil)
	if code != 200 || sum["total_checks"].(float64) < 1 {
		t.Errorf("summary failed: %v", sum)
	}
	// audit
	svc.ReconcileAll(t.Context(), pid)
	code, audit := doJSON(t, h, "GET", "/api/projects/"+pid+"/audit", nil)
	if code != 200 || audit["consistent"] != true {
		t.Errorf("audit not consistent: %v", audit)
	}
}

func TestFrontendServed(t *testing.T) {
	_, h := newAPI(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200 for /, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("结构荷载")) {
		t.Error("frontend HTML should contain the page title")
	}
}
