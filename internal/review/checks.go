package review

import (
	"fmt"

	"task133-structload/internal/model"
)

type checkCounts struct {
	Evaluated int
	Pass      int
	Marginal  int
	Fail      int
}

// checkStatusFindings promotes governing engineering results into the review
// report and catches the special case where a calculation was emitted without
// any usable member capacity.
func checkStatusFindings(checks []model.Check) ([]Finding, checkCounts) {
	counts := checkCounts{Evaluated: len(checks)}
	findings := make([]Finding, 0)
	for _, check := range checks {
		switch check.Status {
		case model.StatusPass:
			counts.Pass++
		case model.StatusMarginal:
			counts.Marginal++
			findings = append(findings, Finding{
				Severity:      SeverityWarning,
				Code:          "marginal_utilization",
				ComponentID:   check.ComponentID,
				CombinationID: check.CombinationID,
				Message:       fmt.Sprintf("构件利用率为 %.2f，处于临界区间", float64(check.UR)/100),
			})
		case model.StatusFail:
			counts.Fail++
			findings = append(findings, Finding{
				Severity:      SeverityError,
				Code:          "failed_utilization",
				ComponentID:   check.ComponentID,
				CombinationID: check.CombinationID,
				Message:       fmt.Sprintf("构件利用率为 %.2f，超过承载能力", float64(check.UR)/100),
			})
		}
		if check.CapacityMoment == 0 && check.CapacityShear == 0 && check.CapacityAxial == 0 {
			findings = append(findings, Finding{
				Severity:      SeverityError,
				Code:          "missing_member_capacity",
				ComponentID:   check.ComponentID,
				CombinationID: check.CombinationID,
				Message:       "检查结果没有可用的构件承载力，不能作为合规结论",
			})
		}
	}
	return findings, counts
}
