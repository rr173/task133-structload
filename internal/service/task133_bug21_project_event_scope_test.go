package service

import (
	"context"
	"testing"
)

func TestBug21_ProjectLifecycleEventsStayScopedAndTyped(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	a, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-21A", Name: "a", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	if _, err = svc.CreateProject(ctx, CreateProjectInput{Code: "P-21B", Name: "b", Exposure: "C", Importance: "II", WindSpeed: 380}); err != nil { t.Fatal(err) }
	if _, err = svc.CreateLevel(ctx, a.ID, CreateLevelInput{Name: "L1", GrossArea: 10000}); err != nil { t.Fatal(err) }
	events, err := svc.ListEvents(ctx, a.ID)
	if err != nil || len(events) != 2 { t.Fatalf("events = %#v, %v", events, err) }
	if events[0].Seq != 1 || events[1].Seq != 2 || events[1].EventType != "add_level" { t.Fatalf("project lifecycle = %#v", events) }
}
