package review

import "task133-structload/internal/model"

// Input is the complete, read-only project snapshot required for a review.
// The service supplies it from the same SQLite records used by derivation and
// sufficiency checks, so the report cannot silently inspect stale copies.
type Input struct {
	ProjectID    string
	Components   []model.Component
	Combinations []model.LoadCombination
	ManualCases  []model.LoadCase
	Checks       []model.Check
}

func checkKey(componentID, combinationID string) string {
	return componentID + "\x00" + combinationID
}
