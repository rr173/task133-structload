package review

// Build creates a review report from one consistent project snapshot. The
// report does not mutate load cases or checks; callers can safely request it
// during a design review without triggering a recalculation.
func Build(in Input) Report {
	report := Report{
		ProjectID:    in.ProjectID,
		Components:   len(in.Components),
		Combinations: len(in.Combinations),
	}

	report.Findings = append(report.Findings, combinationFindings(in.Combinations)...)
	report.Findings = append(report.Findings, manualLoadFindings(in.Components, in.ManualCases)...)
	coverage, expected := checkCoverageFindings(in.Components, in.Combinations, in.Checks)
	report.ExpectedChecks = expected
	report.Findings = append(report.Findings, coverage...)
	statusFindings, counts := checkStatusFindings(in.Checks)
	report.EvaluatedChecks = counts.Evaluated
	report.PassCount = counts.Pass
	report.MarginalCount = counts.Marginal
	report.FailCount = counts.Fail
	report.Findings = append(report.Findings, statusFindings...)

	sortFindings(report.Findings)
	report.Ready = !report.hasBlockingFinding()
	return report
}
