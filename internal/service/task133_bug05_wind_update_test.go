package service

import (
	"context"
	"testing"

	"task133-structload/internal/model"
)

func TestBug05_WindUpdateRefreshesDerivedPressure(t *testing.T) {
	svc, _ := newSvc(t); ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "wind", Name: "Wind", Exposure: "C", Importance: "II", WindSpeed: 400})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", Elevation: 12000, GrossArea: 400000}); if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "W-5", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2400, KLL: 2, NominalMoment: 1000000, NominalShear: 1000000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward"}); if err != nil { t.Fatal(err) }
	if err := svc.Derive(ctx, p.ID); err != nil { t.Fatal(err) }
	before, err := svc.ListDerivedCases(ctx, c.ID); if err != nil { t.Fatal(err) }
	windBefore := int64(0); for _, lc := range before { if lc.Kind == model.LoadWind { windBefore = lc.Magnitude } }
	if windBefore == 0 { t.Fatalf("expected initial wind pressure: %#v", before) }
	updatedSpeed := int64(600)
	if _, err := svc.UpdateProject(ctx, p.ID, UpdateProjectInput{WindSpeed: &updatedSpeed}); err != nil { t.Fatal(err) }
	stored, err := svc.GetProject(ctx, p.ID); if err != nil || stored.WindSpeed != updatedSpeed { t.Fatalf("wind speed was not persisted: %v %#v", err, stored) }
	if _, err := svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	after, err := svc.ListDerivedCases(ctx, c.ID); if err != nil { t.Fatal(err) }
	windAfter := int64(0); for _, lc := range after { if lc.Kind == model.LoadWind { windAfter = lc.Magnitude } }
	if windAfter == 0 || windAfter == windBefore { t.Fatalf("updated wind speed must regenerate pressure: before=%d after=%d", windBefore, windAfter) }
}
