package service

import (
	"context"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

func newSvc(t *testing.T) (*Service, *store.Store) {
	t.Helper()
	s, err := store.Open(t.TempDir() + "/svc.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return New(s, clock.NewFake(time.UnixMilli(1740000000000))), s
}

func TestCreateProjectValidation(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	// missing exposure
	if _, err := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Importance: "II", WindSpeed: 400, GroundSnow: 0}); err == nil {
		t.Fatal("expected error for missing exposure")
	}
	// invalid importance
	if _, err := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "V", WindSpeed: 400, GroundSnow: 0}); err == nil {
		t.Fatal("expected error for invalid importance")
	}
	// zero wind speed
	if _, err := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 0, GroundSnow: 0}); err == nil {
		t.Fatal("expected error for zero wind speed")
	}
	// valid
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	if err != nil {
		t.Fatal(err)
	}
	if p.ID == "" || p.Kd != 85 {
		t.Errorf("unexpected project: %+v", p)
	}
}

func TestCreateComponentAndLoadCase(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 10000, NominalShear: 5000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward", RoofSlope: 0})
	if err != nil {
		t.Fatal(err)
	}
	if c.ProjectID != p.ID {
		t.Errorf("component project should resolve from level: %s", c.ProjectID)
	}
	lc, err := svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 240, Direction: "gravity"})
	if err != nil {
		t.Fatal(err)
	}
	if lc.Origin != model.OriginManual {
		t.Error("manual case should have manual origin")
	}
	// reject W/S manual
	if _, err := svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "W", LoadType: "area_pressure", Magnitude: 80, Direction: "lateral"}); err == nil {
		t.Fatal("should reject manual W case")
	}
}

func TestDeriveProducesWindSnowLive(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 10000, NominalShear: 5000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward", RoofSlope: 0})
	svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 240, Direction: "gravity"})
	if err := svc.Derive(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	derived, _ := svc.ListDerivedCases(ctx, c.ID)
	kinds := map[string]bool{}
	for _, d := range derived {
		kinds[string(d.Kind)] = true
	}
	if !kinds["W"] || !kinds["S"] || !kinds["L"] {
		t.Errorf("derive should produce W, S, L: %+v", kinds)
	}
}

func TestRunChecksAndSummary(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward", RoofSlope: 0})
	svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity"})
	svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 240, Direction: "gravity"})
	svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "1.2D+1.6L+0.5S", Kind: "strength", CoeffD: 120, CoeffL: 160, CoeffS: 50})
	res, err := svc.RunChecks(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.ChecksCount < 1 {
		t.Errorf("expected >=1 check, got %d", res.ChecksCount)
	}
	sum, err := svc.Summary(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sum.TotalChecks < 1 {
		t.Errorf("summary total = %d", sum.TotalChecks)
	}
}

func TestReconcileAllIdempotent(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward", RoofSlope: 0})
	svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity"})
	svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "c", Kind: "strength", CoeffD: 120})
	svc.RunChecks(ctx, p.ID)
	before, _ := svc.ListChecksByComponent(ctx, c.ID)
	svc.ReconcileAll(ctx, p.ID)
	after, _ := svc.ListChecksByComponent(ctx, c.ID)
	if len(before) != len(after) {
		t.Fatalf("check count mismatch: %d vs %d", len(before), len(after))
	}
	for i := range before {
		if before[i].UR != after[i].UR || before[i].Status != after[i].Status {
			t.Errorf("check %d diverged: before=%+v after=%+v", i, before[i], after[i])
		}
	}
}

func TestOverrideRecomputes(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 0})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 10000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none", RoofSlope: 0})
	svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity"})
	svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "c", Kind: "strength", CoeffD: 120})
	svc.RunChecks(ctx, p.ID)
	before, _ := svc.ListChecksByComponent(ctx, c.ID)
	beforeUR := before[0].UR
	if _, err := svc.OverrideComponent(ctx, c.ID, OverrideComponentInput{Field: "nominal_moment", NewValue: "200000", Reason: "r"}); err != nil {
		t.Fatal(err)
	}
	after, _ := svc.ListChecksByComponent(ctx, c.ID)
	if after[0].UR >= beforeUR {
		t.Errorf("UR should drop after Mn increase: %d -> %d", beforeUR, after[0].UR)
	}
}

func TestOverrideRejectsInvalidField(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, CreateProjectInput{Code: "x", Name: "x", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 0})
	lvl, _ := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, KLL: 2, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none", TributaryArea: 400000, TributaryWidth: 2500})
	if _, err := svc.OverrideComponent(ctx, c.ID, OverrideComponentInput{Field: "id", NewValue: "hax", Reason: "r"}); err == nil {
		t.Fatal("should reject non-overrideable field")
	}
}
