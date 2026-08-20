package review

import (
	"testing"

	"task133-structload/internal/model"
)

func TestBug16_ReviewShowsEveryMissingCheckBeforeInputWarnings(t *testing.T) {
	report := Build(Input{
		ProjectID: "P-16",
		Components: []model.Component{{ID: "C1", Code: "B1"}, {ID: "C2", Code: "B2"}},
		Combinations: []model.LoadCombination{{ID: "LC1", Name: "strength"}},
		ManualCases: []model.LoadCase{{ComponentID: "C1", Origin: model.OriginManual}, {ComponentID: "C2", Origin: model.OriginDerived}},
	})
	if len(report.Findings) != 3 { t.Fatalf("findings = %#v", report.Findings) }
	if report.Findings[0].Code != "missing_check" || report.Findings[1].Code != "missing_check" { t.Fatalf("blocking findings must lead: %#v", report.Findings) }
	if report.Findings[2].Code != "component_without_manual_load" || report.Findings[2].ComponentID != "C2" { t.Fatalf("manual input warning = %#v", report.Findings) }
}
