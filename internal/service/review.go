package service

import (
	"context"

	"task133-structload/internal/model"
	"task133-structload/internal/review"
)

// Review builds a read-only structural design-review report from the same
// persisted inputs and check rows that the normal calculation workflow uses.
// It does not derive or recalculate: callers can distinguish an incomplete
// calculation run from one that needs an explicit recomputation.
func (svc *Service) Review(ctx context.Context, projectID string) (review.Report, error) {
	if _, err := svc.GetProject(ctx, projectID); err != nil {
		return review.Report{}, wrapErr(err, "get project for review")
	}
	components, err := svc.store.ListComponentsByProject(ctx, projectID)
	if err != nil {
		return review.Report{}, wrapErr(err, "list review components")
	}
	combinations, err := svc.store.ListCombinationsByProject(ctx, projectID)
	if err != nil {
		return review.Report{}, wrapErr(err, "list review combinations")
	}
	checks, err := svc.store.ListChecksByProject(ctx, projectID)
	if err != nil {
		return review.Report{}, wrapErr(err, "list review checks")
	}
	snapshot, err := svc.store.LoadAll(ctx)
	if err != nil {
		return review.Report{}, wrapErr(err, "load review inputs")
	}
	manualCases := make([]model.LoadCase, 0)
	for _, loadCase := range snapshot.ManualCases {
		if loadCase.ProjectID == projectID {
			manualCases = append(manualCases, loadCase)
		}
	}
	return review.Build(review.Input{
		ProjectID:    projectID,
		Components:   components,
		Combinations: combinations,
		ManualCases:  manualCases,
		Checks:       checks,
	}), nil
}
