package service

import (
	"context"
	"testing"

	"task133-structload/internal/model"
)

func TestBug02_RoofUpliftWindKeepsDirectionAndMagnitude(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "uplift", Name: "Uplift", Exposure: "C", Importance: "II", WindSpeed: 400})
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "roof", Elevation: 12000, GrossArea: 800000})
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "R-1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2400, KLL: 2, NominalMoment: 1000000, NominalShear: 1000000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "roof_uplift"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Derive(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	derived, err := svc.ListDerivedCases(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, lc := range derived {
		if lc.Kind == model.LoadWind && lc.Origin == model.OriginDerived {
			if lc.Direction != model.DirUplift || lc.Magnitude != -61 {
				t.Fatalf("roof uplift must remain upward with the calculated pressure: %#v", lc)
			}
			return
		}
	}
	t.Fatalf("expected a derived roof-uplift wind load, got %#v", derived)
}
