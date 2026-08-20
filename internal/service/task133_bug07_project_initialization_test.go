package service

import (
	"context"
	"testing"
)

func TestBug07_ProjectInitializationKeepsDesignInputsAndTrace(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-07", Name: "snow hall", Site: "north", Exposure: "C", Importance: "II", WindSpeed: 380, GroundSnow: 95})
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := svc.GetProject(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.GroundSnow != 95 {
		t.Fatalf("ground snow = %d, want 95", persisted.GroundSnow)
	}
	if _, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "roof", Elevation: 12000, GrossArea: 900000}); err != nil {
		t.Fatal(err)
	}
	events, err := svc.ListEvents(ctx, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %#v", events)
	}
	if events[0].EventType != "create_project" {
		t.Fatalf("first event type = %q", events[0].EventType)
	}
	if events[0].Seq != 1 || events[1].Seq != 2 {
		t.Fatalf("project event sequence = %d,%d, want 1,2", events[0].Seq, events[1].Seq)
	}
}
