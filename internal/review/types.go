// Package review turns persisted structural checks into a deterministic
// design-review report. It deliberately keeps review findings separate from
// stored checks: a finding is an explanation of incomplete evidence or a
// governing result, not another authoritative engineering input.
package review

// Severity expresses how urgently a reviewer must act on a finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Finding identifies one reviewable condition in a project.
type Finding struct {
	Severity      Severity `json:"severity"`
	Code          string   `json:"code"`
	ComponentID   string   `json:"component_id,omitempty"`
	CombinationID string   `json:"combination_id,omitempty"`
	Message       string   `json:"message"`
}

// Report is a read-only review of the inputs and computed checks for a
// project. Ready means that all expected checks exist and no failed or
// unassessed capacity remains.
type Report struct {
	ProjectID       string    `json:"project_id"`
	Components      int       `json:"components"`
	Combinations    int       `json:"combinations"`
	ExpectedChecks  int       `json:"expected_checks"`
	EvaluatedChecks int       `json:"evaluated_checks"`
	PassCount       int       `json:"pass_count"`
	MarginalCount   int       `json:"marginal_count"`
	FailCount       int       `json:"fail_count"`
	Ready           bool      `json:"ready"`
	Findings        []Finding `json:"findings"`
}

func (r Report) hasBlockingFinding() bool {
	for _, finding := range r.Findings {
		if finding.Severity == SeverityError {
			return true
		}
	}
	return false
}
