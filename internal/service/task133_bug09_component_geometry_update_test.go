package service

import (
	"context"
	"strings"
	"testing"
)

func TestBug09_SpanRevisionPreservesGeometryAndTrace(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-09", Name: "span revision", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", GrossArea: 500000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 200000, TributaryWidth: 2400, KLL: 1, NominalMoment: 100000, NominalShear: 100000, PhiB: 90, PhiC: 90, PhiV: 90, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	span := int64(8400)
	updated, err := svc.UpdateComponent(ctx, c.ID, UpdateComponentInput{Span: &span})
	if err != nil { t.Fatal(err) }
	if updated.Span != span { t.Fatalf("returned span = %d, want %d", updated.Span, span) }
	persisted, err := svc.GetComponent(ctx, c.ID)
	if err != nil || persisted.Span != span { t.Fatalf("persisted component = %#v, %v", persisted, err) }
	events, err := svc.ListEvents(ctx, p.ID)
	if err != nil || len(events) < 4 || !strings.Contains(events[len(events)-1].Payload, "8400") { t.Fatalf("update event = %#v, %v", events, err) }
}
