package service

import (
	"context"
	"strings"
	"testing"
)

func TestBug08_ManualLateralLoadKeepsItsEngineeringMeaning(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-08", Name: "braced frame", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", Elevation: 0, GrossArea: 500000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "C1", Type: "column", Span: 4000, TributaryArea: 10000, TributaryWidth: 1000, KLL: 1, NominalMoment: 100000, NominalAxial: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "E", LoadType: "point_load", Magnitude: 725, Direction: "lateral", Note: "collector force"}); err != nil { t.Fatal(err) }
	loads, err := svc.ListLoadCases(ctx, c.ID)
	if err != nil || len(loads) != 1 { t.Fatalf("loads = %#v, %v", loads, err) }
	if loads[0].Magnitude != 725 || string(loads[0].Direction) != "lateral" { t.Fatalf("stored load = %#v", loads[0]) }
	events, err := svc.ListEvents(ctx, p.ID)
	if err != nil { t.Fatal(err) }
	if len(events) < 3 || !strings.Contains(events[len(events)-1].Payload, "collector force") { t.Fatalf("load event = %#v", events) }
}
