package review

import (
	"fmt"

	"task133-structload/internal/model"
)

// manualLoadFindings verifies that every included component has an
// authoritative manual input. Derived wind, snow, and reduced live load are
// intentionally excluded: they are outputs and cannot provide design input.
func manualLoadFindings(components []model.Component, manualCases []model.LoadCase) []Finding {
	manualByComponent := make(map[string]int, len(components))
	for _, loadCase := range manualCases {
		if loadCase.Origin == model.OriginManual {
			manualByComponent[loadCase.ComponentID]++
		}
	}
	findings := make([]Finding, 0)
	for _, component := range components {
		if manualByComponent[component.ID] != 0 {
			continue
		}
		findings = append(findings, Finding{
			Severity:    SeverityWarning,
			Code:        "component_without_manual_load",
			ComponentID: component.ID,
			Message:     fmt.Sprintf("构件 %s 没有手工荷载工况；派生荷载不能替代设计输入", component.Code),
		})
	}
	return findings
}
