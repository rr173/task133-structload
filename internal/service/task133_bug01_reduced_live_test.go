package service

import (
	"context"
	"testing"

	"task133-structload/internal/loadcode"
	"task133-structload/internal/model"
)

func TestBug01_ReducedLiveLoadControlsCombination(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "large-bay", Name: "Large bay", Exposure: "C", Importance: "II", WindSpeed: 400})
	if err != nil {
		t.Fatal(err)
	}
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", Elevation: 12000, GrossArea: 1200000})
	if err != nil {
		t.Fatal(err)
	}
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B-1", Type: "beam", Span: 6000, TributaryArea: 900000, TributaryWidth: 2500, KLL: 2, NominalMoment: 1000000, NominalShear: 1000000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 240, Direction: "gravity"}); err != nil {
		t.Fatal(err)
	}
	combo, err := svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "L", Kind: "strength", CoeffL: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RunChecks(ctx, p.ID); err != nil {
		t.Fatal(err)
	}
	derived, err := svc.ListDerivedCases(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	var reduced model.LoadCase
	for _, lc := range derived {
		if lc.Kind == model.LoadLive && lc.Origin == model.OriginDerived {
			reduced = lc
		}
	}
	if reduced.Magnitude == 0 || reduced.Magnitude >= 240 {
		t.Fatalf("expected a reduced live-load case, got %#v", derived)
	}
	checks, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(checks) != 1 {
		t.Fatalf("expected one check, err=%v checks=%#v", err, checks)
	}
	var want loadcode.DemandEffect
	want.AddComponentDemand(c, loadcode.EffectiveMagnitude{Kind: model.LoadLive, LoadType: reduced.LoadType, Magnitude: reduced.Magnitude, Direction: reduced.Direction, CoefCenti: combo.CoeffL})
	if got := checks[0]; got.DemandMoment != want.Moment || got.DemandShear != want.Shear || got.DemandAxial != want.Axial {
		t.Fatalf("combination did not use the reduced live load: got M/V/P=%d/%d/%d want=%d/%d/%d", got.DemandMoment, got.DemandShear, got.DemandAxial, want.Moment, want.Shear, want.Axial)
	}
}
