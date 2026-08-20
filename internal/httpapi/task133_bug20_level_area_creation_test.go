package httpapi

import (
	"context"
	"net/http"
	"testing"
)

func TestBug20_LevelGrossAreaSurvivesCreation(t *testing.T) {
	svc, h := newAPI(t)
	_, project := doJSON(t, h, http.MethodPost, "/api/projects", map[string]any{"code": "P-20", "name": "floor", "exposure": "C", "importance": "II", "wind_speed": 380})
	id := project["id"].(string)
	code, level := doJSON(t, h, http.MethodPost, "/api/projects/"+id+"/levels", map[string]any{"name": "L2", "elevation": 3000, "gross_area": 880000})
	if code != http.StatusCreated || level["gross_area"] != float64(880000) { t.Fatalf("create level = %d %#v", code, level) }
	levels, err := svc.ListLevels(context.Background(), id)
	if err != nil || len(levels) != 1 || levels[0].GrossArea != 880000 { t.Fatalf("stored levels = %#v, %v", levels, err) }
}
