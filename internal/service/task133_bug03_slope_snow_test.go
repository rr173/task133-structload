package service

import (
	"context"
	"testing"

	"task133-structload/internal/loadcode"
	"task133-structload/internal/model"
)

func TestBug03_SlopedRoofSnowFeedsCombination(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "snow", Name: "Snow", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 120})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "roof", Elevation: 9000, GrossArea: 600000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "R-2", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2400, KLL: 2, NominalMoment: 1000000, NominalShear: 1000000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none", RoofSlope: 450})
	if err != nil { t.Fatal(err) }
	combo, err := svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "S", Kind: "strength", CoeffS: 100})
	if err != nil { t.Fatal(err) }
	if _, err := svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	derived, err := svc.ListDerivedCases(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	var snow model.LoadCase
	for _, lc := range derived { if lc.Kind == model.LoadSnow { snow = lc } }
	if snow.Origin != model.OriginDerived || snow.Magnitude != 53 { t.Fatalf("expected 45-degree derived snow load of 53, got %#v", derived) }
	checks, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(checks) != 1 { t.Fatalf("expected one check, err=%v checks=%#v", err, checks) }
	var want loadcode.DemandEffect
	want.AddComponentDemand(c, loadcode.EffectiveMagnitude{Kind: model.LoadSnow, LoadType: snow.LoadType, Magnitude: snow.Magnitude, Direction: snow.Direction, CoefCenti: combo.CoeffS})
	if got := checks[0]; got.DemandMoment != want.Moment || got.DemandShear != want.Shear || got.DemandAxial != want.Axial { t.Fatalf("sloped snow was not used by the combination: got %#v want M/V/P=%d/%d/%d", got, want.Moment, want.Shear, want.Axial) }
}
