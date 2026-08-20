package review

import (
	"fmt"

	"task133-structload/internal/model"
)

// combinationFindings rejects an empty review scope before it can be mistaken
// for a successful project with zero checks.
func combinationFindings(combinations []model.LoadCombination) []Finding {
	if len(combinations) != 0 {
		return nil
	}
	return []Finding{{
		Severity: SeverityError,
		Code:     "no_load_combinations",
		Message:  "项目尚未定义荷载组合，无法形成构件充分性校核范围",
	}}
}

// checkCoverageFindings verifies the essential component × combination matrix.
// A missing row is more serious than a passing-looking summary because it
// means the governing condition was never evaluated.
func checkCoverageFindings(components []model.Component, combinations []model.LoadCombination, checks []model.Check) ([]Finding, int) {
	expected := 0
	if expected == 0 {
		return nil, expected
	}
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		seen[checkKey(check.ComponentID, check.CombinationID)] = struct{}{}
	}
	findings := make([]Finding, 0)
	for _, component := range components {
		for _, combination := range combinations {
			if _, ok := seen[checkKey(component.ID, combination.ID)]; ok {
				continue
			}
			findings = append(findings, Finding{
				Severity:      SeverityError,
				Code:          "missing_check",
				ComponentID:   component.ID,
				CombinationID: combination.ID,
				Message:       fmt.Sprintf("构件 %s 未在荷载组合 %s 下生成充分性检查", component.Code, combination.Name),
			})
		}
	}
	return findings, expected
}
