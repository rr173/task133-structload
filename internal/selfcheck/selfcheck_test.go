package selfcheck

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/service"
	"task133-structload/internal/store"
)

// TestRunSmoke invokes the same scenarios the CLI --smoke-test runs.
func TestRunSmoke(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "selfcheck.db")
	out, err := Run(path)
	if err != nil {
		t.Fatalf("smoke failed: %v\n%s", err, out)
	}
	// The smoke run must create the DB file.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected db file to exist: %v", err)
	}
}

// TestExpectedHelpers sanity-checks the independent reference math used to
// cross-validate loadcode so a silent drift is caught.
func TestExpectedHelpers(t *testing.T) {
	// 40 m², KLL=2, Lo=240 -> reduced L should be strictly less than Lo.
	if got := expectedReducedLiveLoad(2, 400000, 240); got <= 0 || got >= 240 {
		t.Errorf("reduced L out of range: %d", got)
	}
	// 5 m² -> no reduction.
	if got := expectedReducedLiveLoad(2, 50000, 240); got != 240 {
		t.Errorf("small area should be Lo=240: %d", got)
	}
	// windward wind pressure positive, leeward negative.
	ww := expectedWindPressureCenti(10000, 100, 85, 400, "C", "windward", 85)
	lw := expectedWindPressureCenti(10000, 100, 85, 400, "C", "leeward", 85)
	if ww <= 0 {
		t.Errorf("windward wind should be > 0: %d", ww)
	}
	if lw >= 0 {
		t.Errorf("leeward wind should be < 0: %d", lw)
	}
}

// TestReconcileIdempotentService verifies ReconcileAll twice produces identical checks.
func TestReconcileIdempotentService(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "rec.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	svc := service.New(s, clock.NewFake(time.UnixMilli(1740000000000)))
	ctx := context.Background()
	p, _ := svc.CreateProject(ctx, service.CreateProjectInput{Code: "X", Name: "X", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 95})
	lvl, _ := svc.CreateLevel(ctx, p.ID, service.CreateLevelInput{Name: "1F", Elevation: 10000, GrossArea: 1000000})
	c, _ := svc.CreateComponent(ctx, "", service.CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "windward", RoofSlope: 0})
	svc.AddManualLoadCase(ctx, c.ID, service.AddManualLoadCaseInput{Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity"})
	svc.CreateCombination(ctx, p.ID, service.CreateCombinationInput{Name: "c", Kind: "strength", CoeffD: 120, CoeffL: 160, CoeffS: 50})
	svc.RunChecks(ctx, p.ID)
	first, _ := svc.ListChecksByComponent(ctx, c.ID)
	svc.ReconcileAll(ctx, p.ID)
	svc.ReconcileAll(ctx, p.ID)
	second, _ := svc.ListChecksByComponent(ctx, c.ID)
	if len(first) != len(second) {
		t.Fatalf("count drift: %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].UR != second[i].UR || first[i].Status != second[i].Status {
			t.Errorf("reconcile not idempotent: %+v vs %+v", first[i], second[i])
		}
	}
}
