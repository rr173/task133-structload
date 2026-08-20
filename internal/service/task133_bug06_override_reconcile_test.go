package service

import (
	"context"
	"testing"
)

func TestBug06_ApprovedCapacityOverrideStaysConsistent(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "ovr", Name: "override", Exposure: "C", Importance: "II", WindSpeed: 400})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", Elevation: 10000, GrossArea: 1000000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 10000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity"}); err != nil { t.Fatal(err) }
	if _, err = svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "D", Kind: "strength", CoeffD: 120}); err != nil { t.Fatal(err) }
	if _, err = svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	before, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(before) != 1 { t.Fatalf("before checks: %#v, %v", before, err) }
	if _, err = svc.OverrideComponent(ctx, c.ID, OverrideComponentInput{Field: "nominal_moment", NewValue: "200000", Reason: "approved reinforcement"}); err != nil { t.Fatal(err) }
	persisted, err := svc.GetComponent(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	if persisted.NominalMoment != 200000 { t.Fatalf("persisted nominal moment = %d, want 200000", persisted.NominalMoment) }
	immediate, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(immediate) != 1 { t.Fatalf("immediate checks: %#v, %v", immediate, err) }
	if immediate[0].UR >= before[0].UR { t.Fatalf("immediate utilization did not drop: %d -> %d", before[0].UR, immediate[0].UR) }
	if err = svc.ReconcileAll(ctx, p.ID); err != nil { t.Fatal(err) }
	recovered, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(recovered) != 1 { t.Fatalf("recovered checks: %#v, %v", recovered, err) }
	if recovered[0].UR != immediate[0].UR { t.Fatalf("recovered utilization = %d, immediate = %d", recovered[0].UR, immediate[0].UR) }
}
