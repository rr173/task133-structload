package service

import (
	"context"
	"fmt"

	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// Derive refreshes all derived load cases (W/S/reduced-L) for a project.
// Clears existing derived cases first, then recomputes from authoritative inputs.
func (svc *Service) Derive(ctx context.Context, projectID string) error {
	eventID := idlib.NewID("drv")
	err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		// Clear existing derived cases for this project.
		if _, err := tx.ExecContext(ctx, `DELETE FROM load_cases WHERE project_id=? AND origin='derived'`, projectID); err != nil {
			return fmt.Errorf("delete derived load_cases: %w", err)
		}
		p, err := svc.store.GetProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		comps, err := svc.componentsForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		var allCases []model.LoadCase
		for _, c := range comps {
			derived, err := svc.deriveForComponent(ctx, tx, p, c, eventID)
			if err != nil {
				return err
			}
			allCases = append(allCases, derived...)
		}
		if err := svc.store.CreateLoadCasesInTx(ctx, tx, allCases); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return wrapErr(err, "derive")
	}
	return svc.appendEvent(ctx, projectID, "derive", []byte(`{"event_id":"`+eventID+`"}`))
}

// ReconcileAll is the restart recovery path: it wipes derived load cases and
// checks, re-derives from authoritative inputs, re-applies overrides, and
// recomputes all checks. The result is byte-identical to the pre-restart state
// for the same authoritative inputs.
func (svc *Service) ReconcileAll(ctx context.Context, projectID string) error {
	err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		// 1. Wipe derived state.
		if _, err := tx.ExecContext(ctx, `DELETE FROM checks WHERE project_id=?`, projectID); err != nil {
			return fmt.Errorf("delete checks: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM load_cases WHERE project_id=? AND origin='derived'`, projectID); err != nil {
			return fmt.Errorf("delete derived load_cases: %w", err)
		}
		p, err := svc.store.GetProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		// 2. Re-derive load cases.
		comps, err := svc.componentsForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		// 3. Re-apply overrides to component authoritative fields (last value per field wins).
		overrides, err := svc.overridesForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		comps = applyOverrides(comps, overrides)
		var allCases []model.LoadCase
		for _, c := range comps {
			derived, err := svc.deriveForComponent(ctx, tx, p, c, "reconcile")
			if err != nil {
				return err
			}
			allCases = append(allCases, derived...)
		}
		if err := svc.store.CreateLoadCasesInTx(ctx, tx, allCases); err != nil {
			return err
		}
		// 4. Recompute checks (use overridden components).
		combos, err := svc.combosForProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		for _, c := range comps {
			for _, combo := range combos {
				chk, err := svc.computeCheck(ctx, tx, p, c, combo, "reconcile")
				if err != nil {
					return err
				}
				if err := svc.store.UpsertCheckInTx(ctx, tx, chk); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return wrapErr(err, "reconcile")
	}
	return svc.appendEvent(ctx, projectID, "reconcile", []byte(`{}`))
}

// overridesForProjectInTx reads overrides for a project within a transaction.
func (svc *Service) overridesForProjectInTx(ctx context.Context, tx store.DBTX, projectID string) ([]model.Override, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,project_id,component_id,field,old_value,new_value,reason,applied_at
		FROM overrides WHERE project_id=? ORDER BY applied_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("query overrides: %w", err)
	}
	defer rows.Close()
	var out []model.Override
	for rows.Next() {
		var o model.Override
		var ms int64
		if err := rows.Scan(&o.ID, &o.ProjectID, &o.ComponentID, &o.Field, &o.OldValue, &o.NewValue, &o.Reason, &ms); err != nil {
			return nil, err
		}
		o.AppliedAt = fromMS(ms)
		out = append(out, o)
	}
	return out, rows.Err()
}

// applyOverrides returns components with the last override value per (component, field) applied.
// Only a small whitelist of fields is overridable.
func applyOverrides(comps []model.Component, overrides []model.Override) []model.Component {
	byID := map[string]*model.Component{}
	for i := range comps {
		byID[comps[i].ID] = &comps[i]
	}
	// overrides are already sorted by applied_at; later values overwrite earlier.
	for _, o := range overrides {
		c, ok := byID[o.ComponentID]
		if !ok {
			continue
		}
		switch o.Field {
		case "nominal_moment":
			c.NominalMoment = mustParseInt(o.NewValue)
		case "nominal_axial":
			c.NominalAxial = mustParseInt(o.NewValue)
		case "nominal_shear":
			c.NominalShear = mustParseInt(o.NewValue)
		case "phi_b":
			c.PhiB = mustParseInt(o.NewValue)
		case "phi_c":
			c.PhiC = mustParseInt(o.NewValue)
		case "phi_v":
			c.PhiV = mustParseInt(o.NewValue)
		case "wind_surface":
			c.WindSurface = model.WindSurface(o.NewValue)
		case "roof_slope":
			c.RoofSlope = mustParseInt(o.NewValue)
		case "tributary_area":
			c.TributaryArea = mustParseInt(o.NewValue)
		case "kll":
			c.KLL = mustParseInt(o.NewValue)
		}
	}
	return comps
}

// mustParseInt is duplicated from store.component to keep reconcile self-contained.
func mustParseInt(s string) int64 {
	v, err := parseSignedInt(s)
	if err != nil {
		return 0
	}
	return v
}
