package review

import (
	"testing"

	"task133-structload/internal/model"
)

func TestBug12_ReviewKeepsMissingManualInputVisibleAfterErrors(t *testing.T) {
	report := Build(Input{
		ProjectID: "P-12",
		Components:   []model.Component{{ID: "C1", Code: "B1"}, {ID: "C2", Code: "B2"}},
		Combinations: []model.LoadCombination{{ID: "LC1", Name: "strength"}},
		ManualCases:  []model.LoadCase{{ComponentID: "C1", Origin: model.OriginManual}, {ComponentID: "C2", Origin: model.OriginDerived}},
		Checks:       []model.Check{{ComponentID: "C1", CombinationID: "LC1", Status: model.StatusPass, CapacityMoment: 100}},
	})
	if len(report.Findings) < 2 { t.Fatalf("findings = %#v", report.Findings) }
	if report.Findings[0].Code != "missing_check" { t.Fatalf("first finding = %#v", report.Findings[0]) }
	seen := false
	for _, finding := range report.Findings {
		if finding.Code == "component_without_manual_load" && finding.ComponentID == "C2" { seen = true }
	}
	if !seen { t.Fatalf("derived load incorrectly satisfied manual-input review: %#v", report.Findings) }
}
