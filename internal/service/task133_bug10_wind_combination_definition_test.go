package service

import (
	"context"
	"strings"
	"testing"
)

func TestBug10_WindCombinationDefinitionRemainsAuthoritative(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-10", Name: "wind combo", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	combo, err := svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "1.2D+1.0W", Kind: "strength", CoeffD: 120, CoeffS: 50, CoeffW: 160})
	if err != nil { t.Fatal(err) }
	if combo.CoeffW != 160 { t.Fatalf("returned wind factor = %d", combo.CoeffW) }
	stored, err := svc.ListCombinations(ctx, p.ID)
	if err != nil || len(stored) != 1 || stored[0].CoeffW != 160 { t.Fatalf("stored combinations = %#v, %v", stored, err) }
	events, err := svc.ListEvents(ctx, p.ID)
	if err != nil || len(events) < 2 || !strings.Contains(events[len(events)-1].Payload, "160") { t.Fatalf("combination event = %#v, %v", events, err) }
}
