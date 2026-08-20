package service

import (
	"context"

	"task133-structload/internal/model"
)

// Summary computes the compliance roll-up for a project from persisted checks.
func (svc *Service) Summary(ctx context.Context, projectID string) (model.Summary, error) {
	checks, err := svc.store.ListChecksByProject(ctx, projectID)
	if err != nil {
		return model.Summary{}, wrapErr(err, "list checks")
	}
	s := model.Summary{ProjectID: projectID, TotalChecks: len(checks)}
	var maxUR int64 = 0
	var firstFailCheck *model.Check
	for i := range checks {
		c := &checks[i]
		s.TotalChecks++
		switch c.Status {
		case model.StatusPass:
			s.PassCount++
		case model.StatusMarginal:
			s.MarginalCount++
		case model.StatusFail:
			s.FailCount++
			if firstFailCheck == nil {
				firstFailCheck = c
			}
		}
		if c.UR > maxUR {
			maxUR = c.UR
		}
	}
	s.MaxUR = maxUR
	if firstFailCheck != nil {
		s.FirstFailComponent = firstFailCheck.ComponentID
		s.FirstFailCombo = firstFailCheck.CombinationID
	}
	return s, nil
}

// ListEvents returns the event log for a project.
func (svc *Service) ListEvents(ctx context.Context, projectID string) ([]model.EventLog, error) {
	return svc.store.ListEvents(ctx, projectID)
}

// Audit recomputes derived cases and checks, then reports any divergence from
// what is currently persisted. Used to verify recovery consistency.
func (svc *Service) Audit(ctx context.Context, projectID string) (model.AuditResult, error) {
	res := model.AuditResult{ProjectID: projectID}
	// Snapshot the current derived cases + checks.
	beforeDerived, err := svc.store.LoadAll(ctx)
	if err != nil {
		return res, wrapErr(err, "load all before")
	}
	beforeChecks, err := svc.store.ListChecksByProject(ctx, projectID)
	if err != nil {
		return res, wrapErr(err, "list checks before")
	}
	// Reconcile (idempotent: same authoritative inputs -> same derived state).
	if err := svc.ReconcileAll(ctx, projectID); err != nil {
		return res, err
	}
	afterDerived, err := svc.store.LoadAll(ctx)
	if err != nil {
		return res, wrapErr(err, "load all after")
	}
	afterChecks, err := svc.store.ListChecksByProject(ctx, projectID)
	if err != nil {
		return res, wrapErr(err, "list checks after")
	}
	res.DerivedCases = len(afterDerived.DerivedCases)
	res.Checks = len(afterChecks)
	res.StaleDerivedCases = countDerivedDiff(beforeDerived.DerivedCases, afterDerived.DerivedCases)
	res.StaleChecks = countCheckDiff(beforeChecks, afterChecks)
	res.Consistent = res.StaleDerivedCases == 0 && res.StaleChecks == 0
	return res, nil
}

// countDerivedDiff counts how many derived cases differ between snapshots.
// Only the computed value (magnitude) matters for staleness; IDs/timestamps
// are regenerated on each ReconcileAll and must not count as divergence.
func countDerivedDiff(a, b []model.LoadCase) int {
	am := derivedMap(a)
	bm := derivedMap(b)
	diff := 0
	for k, va := range am {
		vb, ok := bm[k]
		if !ok || vb.Magnitude != va.Magnitude || vb.LoadType != va.LoadType || vb.Direction != va.Direction {
			diff++
		}
	}
	for k := range bm {
		if _, ok := am[k]; !ok {
			diff++
		}
	}
	return diff
}

func derivedMap(cases []model.LoadCase) map[string]model.LoadCase {
	m := map[string]model.LoadCase{}
	for _, c := range cases {
		// key by component+kind+origin so identical recomputation matches.
		m[c.ComponentID+"|"+string(c.Kind)] = c
	}
	return m
}

// countCheckDiff counts how many checks differ by (component,combination) key or value.
func countCheckDiff(a, b []model.Check) int {
	am := checkMap(a)
	bm := checkMap(b)
	diff := 0
	for k, va := range am {
		if vb, ok := bm[k]; !ok || !checkEqual(va, vb) {
			diff++
		}
	}
	for k := range bm {
		if _, ok := am[k]; !ok {
			diff++
		}
	}
	return diff
}

func checkMap(checks []model.Check) map[string]model.Check {
	m := map[string]model.Check{}
	for _, c := range checks {
		m[c.ComponentID+"|"+c.CombinationID] = c
	}
	return m
}

func checkEqual(a, b model.Check) bool {
	return a.UR == b.UR && a.Status == b.Status &&
		a.DemandMoment == b.DemandMoment && a.DemandShear == b.DemandShear && a.DemandAxial == b.DemandAxial &&
		a.CapacityMoment == b.CapacityMoment && a.CapacityShear == b.CapacityShear && a.CapacityAxial == b.CapacityAxial
}
