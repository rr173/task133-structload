package review

import (
	"testing"

	"task133-structload/internal/model"
)

func TestBug11_ReviewBlocksIncompleteEvidence(t *testing.T) {
	report := Build(Input{
		ProjectID: "P-11",
		Components: []model.Component{{ID: "C1", Code: "B1"}, {ID: "C2", Code: "B2"}},
		Combinations: []model.LoadCombination{{ID: "LC1", Name: "strength"}},
		Checks: []model.Check{{ComponentID: "C1", CombinationID: "LC1", Status: model.StatusPass}},
	})
	if report.Ready { t.Fatal("review must not be ready when engineering evidence is incomplete") }
	seen := map[string]bool{}
	for _, finding := range report.Findings { seen[finding.Code] = true }
	if !seen["missing_check"] || !seen["missing_member_capacity"] { t.Fatalf("findings = %#v", report.Findings) }
}
