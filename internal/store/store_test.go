package store

import (
	"context"
	"testing"
	"time"

	"task133-structload/internal/model"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndLoadProject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	p := model.Project{
		ID: "prj-1", Code: "B1", Name: "Tower", Site: "here",
		WindZone: "I", SnowZone: "I", Exposure: "C", Importance: "II",
		WindSpeed: 400, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 95,
		SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI",
		CreatedAt: time.UnixMilli(1740000000000),
	}
	if err := s.CreateProject(ctx, p); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetProjectInTx(ctx, s.db, "prj-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Code != "B1" || got.WindSpeed != 400 || got.Exposure != "C" {
		t.Errorf("project mismatch: %+v", got)
	}
}

func TestUpdateProjectNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.UpdateProject(context.Background(), model.Project{ID: "missing"})
	if err != model.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLevelComponentRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.CreateProject(ctx, model.Project{ID: "p1", Code: "x", Name: "x", Exposure: "B", Importance: "II", WindSpeed: 300, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 0, SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI", CreatedAt: time.UnixMilli(1)}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateLevel(ctx, model.Level{ID: "l1", ProjectID: "p1", Name: "1F", Elevation: 0, GrossArea: 100, CreatedAt: time.UnixMilli(1)}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateComponent(ctx, model.Component{ID: "c1", ProjectID: "p1", LevelID: "l1", Code: "B", Type: "beam", Span: 6000, TributaryArea: 400000, TributaryWidth: 2500, KLL: 2, NominalMoment: 10000, WindSurface: "none", CreatedAt: time.UnixMilli(1)}); err != nil {
		t.Fatal(err)
	}
	comps, err := s.ListComponentsByLevel(ctx, "l1")
	if err != nil {
		t.Fatal(err)
	}
	if len(comps) != 1 || comps[0].ID != "c1" {
		t.Fatalf("expected 1 component, got %+v", comps)
	}
}

func TestLoadCaseManualDerived(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	s.CreateProject(ctx, model.Project{ID: "p1", Code: "x", Name: "x", Exposure: "B", Importance: "II", WindSpeed: 300, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 0, SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI", CreatedAt: time.UnixMilli(1)})
	s.CreateComponent(ctx, model.Component{ID: "c1", ProjectID: "p1", LevelID: "", Code: "B", Type: "beam", Span: 6000, WindSurface: "none", CreatedAt: time.UnixMilli(1)})
	s.CreateLoadCase(ctx, model.LoadCase{ID: "lc1", ProjectID: "p1", ComponentID: "c1", Kind: "D", LoadType: "area_pressure", Magnitude: 150, Direction: "gravity", Origin: "manual", CreatedAt: time.UnixMilli(1)})
	s.CreateLoadCase(ctx, model.LoadCase{ID: "lc2", ProjectID: "p1", ComponentID: "c1", Kind: "W", LoadType: "area_pressure", Magnitude: 80, Direction: "lateral", Origin: "derived", CreatedAt: time.UnixMilli(2)})
	snap, err := s.LoadAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.ManualCases) != 1 || snap.ManualCases[0].ID != "lc1" {
		t.Errorf("manual cases: %+v", snap.ManualCases)
	}
	if len(snap.DerivedCases) != 1 || snap.DerivedCases[0].ID != "lc2" {
		t.Errorf("derived cases: %+v", snap.DerivedCases)
	}
}

func TestCheckUpsertIdempotent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	s.CreateProject(ctx, model.Project{ID: "p1", Code: "x", Name: "x", Exposure: "B", Importance: "II", WindSpeed: 300, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 0, SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI", CreatedAt: time.UnixMilli(1)})
	s.CreateCombination(ctx, model.LoadCombination{ID: "cb1", ProjectID: "p1", Name: "c", Kind: "strength"})
	s.CreateComponent(ctx, model.Component{ID: "c1", ProjectID: "p1", Code: "B", Type: "beam", Span: 6000, WindSurface: "none", CreatedAt: time.UnixMilli(1)})
	ch := model.Check{ID: "k1", ProjectID: "p1", CombinationID: "cb1", ComponentID: "c1", UR: 50, Status: "pass", GoverningKind: "moment"}
	if err := s.InTx(ctx, func(tx DBTX) error { return s.UpsertCheckInTx(ctx, tx, ch) }); err != nil {
		t.Fatal(err)
	}
	ch.UR = 99
	ch.Status = "fail"
	if err := s.InTx(ctx, func(tx DBTX) error { return s.UpsertCheckInTx(ctx, tx, ch) }); err != nil {
		t.Fatal(err)
	}
	got, _ := s.ListChecksByComponent(ctx, "c1")
	if len(got) != 1 || got[0].UR != 99 {
		t.Errorf("upsert should replace: %+v", got)
	}
}

func TestEventSeqPerProject(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	s.CreateProject(ctx, model.Project{ID: "p1", Code: "x", Name: "x", Exposure: "B", Importance: "II", WindSpeed: 300, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 0, SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI", CreatedAt: time.UnixMilli(1)})
	if err := s.AppendEvent(ctx, model.EventLog{ID: "e1", ProjectID: "p1", Seq: 1, EventType: "x"}); err != nil {
		t.Fatal(err)
	}
	var seq2 int64
	s.InTx(ctx, func(tx DBTX) error { seq2, _ = s.NextEventSeq(ctx, tx, "p1"); return nil })
	if seq2 != 2 {
		t.Errorf("after 1 append, next seq should be 2, got %d", seq2)
	}
}

func TestDeleteDerivedAndChecks(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	s.CreateProject(ctx, model.Project{ID: "p1", Code: "x", Name: "x", Exposure: "B", Importance: "II", WindSpeed: 300, Kzt: 100, Kd: 85, GustG: 85, GroundSnow: 0, SnowExposureFactor: 100, ThermalFactor: 100, Units: "SI", CreatedAt: time.UnixMilli(1)})
	s.CreateComponent(ctx, model.Component{ID: "c1", ProjectID: "p1", Code: "B", Type: "beam", Span: 6000, WindSurface: "none", CreatedAt: time.UnixMilli(1)})
	s.CreateLoadCase(ctx, model.LoadCase{ID: "d1", ProjectID: "p1", ComponentID: "c1", Kind: "W", Magnitude: 80, Origin: "derived", CreatedAt: time.UnixMilli(1)})
	s.InTx(ctx, func(tx DBTX) error { return s.UpsertCheckInTx(ctx, tx, model.Check{ID: "k1", ProjectID: "p1", CombinationID: "cb1", ComponentID: "c1", UR: 50}) })
	if err := s.DeleteDerivedAndChecks(ctx); err != nil {
		t.Fatal(err)
	}
	snap, _ := s.LoadAll(ctx)
	if len(snap.DerivedCases) != 0 {
		t.Error("derived cases should be cleared")
	}
	chs, _ := s.ListChecksByProject(ctx, "p1")
	if len(chs) != 0 {
		t.Error("checks should be cleared")
	}
}
