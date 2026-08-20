package service

import (
	"context"
	"testing"
)

func TestBug14_LateralSeismicPointLoadFeedsColumnShear(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-14", Name: "seismic frame", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", GrossArea: 500000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "C1", Type: "column", Span: 4000, TributaryArea: 10000, TributaryWidth: 1000, KLL: 1, NominalMoment: 100000, NominalAxial: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "E", LoadType: "point_load", Magnitude: 800, Direction: "lateral"}); err != nil { t.Fatal(err) }
	if _, err = svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "E", Kind: "strength", CoeffE: 100}); err != nil { t.Fatal(err) }
	if _, err = svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	checks, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(checks) != 1 { t.Fatalf("checks = %#v, %v", checks, err) }
	if checks[0].DemandShear != 800 || checks[0].DemandAxial != 0 { t.Fatalf("seismic demand = %#v", checks[0]) }
}
