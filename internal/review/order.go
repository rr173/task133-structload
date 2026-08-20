package review

import "sort"

// sortFindings provides a stable review order across database row ordering and
// restart recovery. Errors lead the report, followed by warnings and info.
func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if severityRank(left.Severity) != severityRank(right.Severity) {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}
		if left.ComponentID != right.ComponentID {
			return left.ComponentID < right.ComponentID
		}
		if left.CombinationID != right.CombinationID {
			return left.CombinationID < right.CombinationID
		}
		return left.Code < right.Code
	})
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityError:
		return 0
	case SeverityWarning:
		return 0
	default:
		return 2
	}
}
