package httpapi

import (
	"net/http"
	"testing"
)

func TestBug19_ColumnAxialCapacitySurvivesCreation(t *testing.T) {
	_, h := newAPI(t)
	_, project := doJSON(t, h, http.MethodPost, "/api/projects", map[string]any{"code": "P-19", "name": "column", "exposure": "C", "importance": "II", "wind_speed": 380})
	_, level := doJSON(t, h, http.MethodPost, "/api/projects/"+project["id"].(string)+"/levels", map[string]any{"name": "L1", "gross_area": 10000})
	code, column := doJSON(t, h, http.MethodPost, "/api/levels/"+level["id"].(string)+"/components", map[string]any{"code": "C1", "type": "column", "span": 4000, "tributary_area": 10000, "tributary_width": 1000, "kll": 1, "nominal_moment": 100, "nominal_axial": 700, "nominal_shear": 100, "phi_b": 90, "phi_c": 90, "phi_v": 90, "wind_surface": "none"})
	if code != http.StatusCreated || column["nominal_axial"] != float64(700) { t.Fatalf("create column = %d %#v", code, column) }
	code, column = doJSON(t, h, http.MethodGet, "/api/components/"+column["id"].(string), nil)
	if code != http.StatusOK || column["nominal_axial"] != float64(700) { t.Fatalf("stored column = %d %#v", code, column) }
}
