package service

import (
	"context"
	"testing"

	"task133-structload/internal/loadcode"
	"task133-structload/internal/model"
)

func TestBug04_SpanUpdateRefreshesPersistedCheck(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "span", Name: "Span", Exposure: "C", Importance: "II", WindSpeed: 400})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", Elevation: 6000, GrossArea: 500000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B-4", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2400, KLL: 2, NominalMoment: 1000000, NominalShear: 1000000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err := svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "line_load", Magnitude: 200, Direction: "gravity"}); err != nil { t.Fatal(err) }
	if _, err := svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "D", Kind: "strength", CoeffD: 100}); err != nil { t.Fatal(err) }
	span := int64(9000)
	if _, err := svc.UpdateComponent(ctx, c.ID, UpdateComponentInput{Span: &span}); err != nil { t.Fatal(err) }
	stored, err := svc.GetComponent(ctx, c.ID)
	if err != nil || stored.Span != span { t.Fatalf("span update was not persisted: %v %#v", err, stored) }
	if _, err := svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	after, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(after) != 1 { t.Fatalf("expected refreshed check: %v %#v", err, after) }
	var want loadcode.DemandEffect
	want.AddComponentDemand(stored, loadcode.EffectiveMagnitude{Kind: model.LoadDead, LoadType: model.LoadLineLoad, Magnitude: 200, Direction: model.DirGravity, CoefCenti: 100})
	if got := after[0]; got.DemandMoment != want.Moment || got.DemandShear != want.Shear || got.DemandAxial != want.Axial { t.Fatalf("check did not use the persisted span: got=%#v want M/V/P=%d/%d/%d", got, want.Moment, want.Shear, want.Axial) }
}
