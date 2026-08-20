package service

import (
	"context"
	"testing"
)

func TestBug15_FailedCheckStaysFailedAcrossSummaryAndReview(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()
	p, err := svc.CreateProject(ctx, CreateProjectInput{Code: "P-15", Name: "status", Exposure: "C", Importance: "II", WindSpeed: 380})
	if err != nil { t.Fatal(err) }
	lvl, err := svc.CreateLevel(ctx, p.ID, CreateLevelInput{Name: "L1", GrossArea: 500000})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateComponent(ctx, "", CreateComponentInput{LevelID: lvl.ID, Code: "B1", Type: "beam", Span: 6000, TributaryArea: 10000, TributaryWidth: 2000, KLL: 1, NominalMoment: 1, NominalShear: 1, PhiB: 100, PhiC: 100, PhiV: 100, WindSurface: "none"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.AddManualLoadCase(ctx, c.ID, AddManualLoadCaseInput{Kind: "D", LoadType: "line_load", Magnitude: 1000, Direction: "gravity"}); err != nil { t.Fatal(err) }
	if _, err = svc.CreateCombination(ctx, p.ID, CreateCombinationInput{Name: "D", Kind: "strength", CoeffD: 100}); err != nil { t.Fatal(err) }
	if _, err = svc.RunChecks(ctx, p.ID); err != nil { t.Fatal(err) }
	checks, err := svc.ListChecksByComponent(ctx, c.ID)
	if err != nil || len(checks) != 1 || string(checks[0].Status) != "fail" { t.Fatalf("checks = %#v, %v", checks, err) }
	summary, err := svc.Summary(ctx, p.ID)
	if err != nil || summary.FailCount != 1 { t.Fatalf("summary = %#v, %v", summary, err) }
	report, err := svc.Review(ctx, p.ID)
	if err != nil || report.Ready { t.Fatalf("review = %#v, %v", report, err) }
}
