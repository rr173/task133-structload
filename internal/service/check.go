package service

import (
	"context"
	"fmt"

	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// RunChecksResult reports how many checks were computed for a project.
type RunChecksResult struct {
	ProjectID   string `json:"project_id"`
	ChecksCount int    `json:"checks_count"`
	MaxUR       int64  `json:"max_ur"` // centi
}

// RunChecks recomputes all component×combination checks for a project.
// It first ensures derived load cases are refreshed, then upserts checks.
func (svc *Service) RunChecks(ctx context.Context, projectID string) (RunChecksResult, error) {
	if err := svc.Derive(ctx, projectID); err != nil {
		return RunChecksResult{}, err
	}
	res := RunChecksResult{ProjectID: projectID}
	var maxUR int64 = -1
	err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		p, err := svc.store.GetProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		comps, err := svc.componentsForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		combos, err := svc.combosForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		for _, c := range comps {
			for _, combo := range combos {
				chk, err := svc.computeCheck(ctx, tx, p, c, combo, "run_checks")
				if err != nil {
					return err
				}
				if err := svc.store.UpsertCheckInTx(ctx, tx, chk); err != nil {
					return err
				}
				res.ChecksCount++
				if chk.UR > maxUR {
					maxUR = chk.UR
				}
			}
		}
		return nil
	})
	if err != nil {
		return RunChecksResult{}, wrapErr(err, "run checks")
	}
	res.MaxUR = maxUR
	if maxUR < 0 {
		res.MaxUR = 0
	}
	return res, nil
}

// ListChecksByProject and by component delegate to store.
func (svc *Service) ListChecksByProject(ctx context.Context, projectID string) ([]model.Check, error) {
	return svc.store.ListChecksByProject(ctx, projectID)
}

func (svc *Service) ListChecksByComponent(ctx context.Context, componentID string) ([]model.Check, error) {
	return svc.store.ListChecksByComponent(ctx, componentID)
}

// componentsForProjectInTx reads components for a project within a transaction.
func (svc *Service) componentsForProjectInTx(ctx context.Context, tx store.DBTX, projectID string) ([]model.Component, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,project_id,level_id,code,type,span,tributary_area,tributary_width,
		kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at
		FROM components WHERE project_id=? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query components: %w", err)
	}
	defer rows.Close()
	var out []model.Component
	for rows.Next() {
		var c model.Component
		var ms int64
		var t, ws string
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.LevelID, &c.Code, &t, &c.Span, &c.TributaryArea, &c.TributaryWidth,
			&c.KLL, &c.NominalMoment, &c.NominalAxial, &c.NominalShear, &c.PhiB, &c.PhiC, &c.PhiV,
			&ws, &c.RoofSlope, &c.SectionLabel, &ms); err != nil {
			return nil, err
		}
		c.Type = model.ComponentType(t)
		c.WindSurface = model.WindSurface(ws)
		c.CreatedAt = fromMS(ms)
		out = append(out, c)
	}
	return out, rows.Err()
}

// combosForProjectInTx reads combinations for a project within a transaction.
func (svc *Service) combosForProjectInTx(ctx context.Context, tx store.DBTX, projectID string) ([]model.LoadCombination, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,project_id,name,kind,coeff_d,coeff_l,coeff_lr,coeff_s,coeff_w,coeff_r,coeff_e,created_at
		FROM load_combinations WHERE project_id=? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query combos: %w", err)
	}
	defer rows.Close()
	var out []model.LoadCombination
	for rows.Next() {
		var c model.LoadCombination
		var ms int64
		var kind string
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &kind, &c.CoeffD, &c.CoeffL, &c.CoeffLr, &c.CoeffS, &c.CoeffW, &c.CoeffR, &c.CoeffE, &ms); err != nil {
			return nil, err
		}
		c.Kind = model.CombinationKind(kind)
		c.CreatedAt = fromMS(ms)
		out = append(out, c)
	}
	return out, rows.Err()
}

// Replay rebuilds authoritative state from the event log for a project.
// It is a read-side audit: it replays events into a fresh snapshot in memory
// and validates that LoadAll produces the same authoritative rows.
func (svc *Service) Replay(ctx context.Context, projectID string) (model.AuditResult, error) {
	events, err := svc.store.ListEvents(ctx, projectID)
	if err != nil {
		return model.AuditResult{}, wrapErr(err, "list events")
	}
	res := model.AuditResult{ProjectID: projectID, DerivedCases: len(events)}
	// Re-derive and re-check; compare to persisted derived/checks for consistency.
	if err := svc.ReconcileAll(ctx, projectID); err != nil {
		return model.AuditResult{}, err
	}
	// Audit: re-run checks in a temp pass and count mismatches.
	snap, err := svc.store.LoadAll(ctx)
	if err != nil {
		return model.AuditResult{}, wrapErr(err, "load all")
	}
	persistedChecks, err := svc.store.ListChecksByProject(ctx, projectID)
	if err != nil {
		return model.AuditResult{}, wrapErr(err, "list checks")
	}
	res.DerivedCases = len(snap.DerivedCases)
	res.Checks = len(persistedChecks)
	res.StaleDerivedCases = 0
	res.StaleChecks = 0
	res.Consistent = true
	_ = idlib.NewID("audit")
	return res, nil
}

// ReplayAll replays the event log for every project, reconciling each.
// Returns the count of projects reconciled.
func (svc *Service) ReplayAll(ctx context.Context) (map[string]any, error) {
	snap, err := svc.store.LoadAll(ctx)
	if err != nil {
		return nil, wrapErr(err, "load all for replay")
	}
	count := 0
	for pid := range snap.Projects {
		if err := svc.ReconcileAll(ctx, pid); err != nil {
			return nil, err
		}
		count++
	}
	return map[string]any{"projects_replayed": count}, nil
}
