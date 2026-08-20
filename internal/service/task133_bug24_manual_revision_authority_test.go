package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

func TestBug24_LatestManualLoadRevisionDrivesDerivationAndChecks(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "revision.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	fake := clock.NewFake(time.UnixMilli(1740000000000))
	svc := New(st, fake)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-24", Name: "revision", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", GrossArea: 100000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 4000, TributaryArea: 400000, TributaryWidth: 2000, KLL: 1, NominalMoment: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 240, Direction: "gravity", Note: "initial"}); err != nil { t.Fatal(err) }
	fake.Advance(time.Second)
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "L", LoadType: "area_pressure", Magnitude: 120, Direction: "gravity", Note: "approved revision"}); err != nil { t.Fatal(err) }
	if err = svc.Derive(ctx, p.ID); err != nil { t.Fatal(err) }
	derived, err := svc.ListDerivedCases(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	var reduced *model.LoadCase
	for i := range derived { if derived[i].Kind == model.LoadLive { reduced = &derived[i]; break } }
	if reduced == nil || reduced.Magnitude != 116 { t.Fatalf("reduced live case = %+v, want revised reduced magnitude 116", reduced) }
	allCases, err := svc.ListLoadCases(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	if len(allCases) < 2 || allCases[0].Origin != model.OriginManual || allCases[0].Magnitude != 120 { t.Fatalf("manual revisions = %+v, want newest first", allCases) }
	allCases, err := svc.ListLoadCases(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	if len(allCases) < 2 || allCases[0].Origin != model.OriginManual || allCases[0].Magnitude != 120 { t.Fatalf("manual revisions = %+v, want newest first", allCases) }
	if _, err = svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "L", Kind: "strength", CoeffL: 100}); err != nil { t.Fatal(err) }
	if _, err = svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	checks, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil { t.Fatal(err) }
	if len(checks) != 1 || checks[0].DemandMoment != 464 || checks[0].DemandShear != 464 { t.Fatalf("check demands = %+v, want reduced revised live demand M/V=464", checks) }
}
