package service

import (
	"context"
	"fmt"

	"task133-structload/internal/idlib"
	"task133-structload/internal/loadcode"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// CreateCombinationInput defines an LRFD factor set.
type CreateCombinationInput struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	CoeffD  int64  `json:"coeff_d"`
	CoeffL  int64  `json:"coeff_l"`
	CoeffLr int64  `json:"coeff_lr"`
	CoeffS  int64  `json:"coeff_s"`
	CoeffW  int64  `json:"coeff_w"`
	CoeffR  int64  `json:"coeff_r"`
	CoeffE  int64  `json:"coeff_e"`
}

// CreateCombination defines a load combination for a project.
func (svc *Service) CreateCombination(ctx context.Context, projectID string, in CreateCombinationInput) (model.LoadCombination, error) {
	if in.Name == "" {
		return model.LoadCombination{}, fmt.Errorf("%w: name required", model.ErrInvalid)
	}
	kind := model.CombinationKind(in.Kind)
	if !kind.Valid() {
		return model.LoadCombination{}, fmt.Errorf("%w: kind must be strength/service", model.ErrInvalid)
	}
	combo := model.LoadCombination{
		ID:        idlib.NewID("cmb"),
		ProjectID: projectID,
		Name:      in.Name,
		Kind:      kind,
		CoeffD:    in.CoeffD,
		CoeffL:    in.CoeffL,
		CoeffLr:   in.CoeffLr,
		CoeffS:    in.CoeffS,
		CoeffW:    in.CoeffW,
		CoeffR:    in.CoeffR,
		CoeffE:    in.CoeffE,
		CreatedAt: svc.now(),
	}
	if !combo.HasAnyNonZero() {
		return model.LoadCombination{}, fmt.Errorf("%w: combination must have at least one non-zero coefficient", model.ErrInvalid)
	}
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		p, err = svc.store.GetProjectInTx(ctx, tx, projectID)
		return err
	}); err != nil {
		return model.LoadCombination{}, wrapErr(err, "get project")
	}
	combo.ProjectID = p.ID
	if err := svc.store.CreateCombination(ctx, combo); err != nil {
		return model.LoadCombination{}, wrapErr(err, "create combination")
	}
	if err := svc.appendEvent(ctx, combo.ProjectID, "define_combination", idlib.MustJSON(combo)); err != nil {
		return model.LoadCombination{}, wrapErr(err, "log define combination")
	}
	return combo, nil
}

// UpdateCombination edits a combination's factors.
func (svc *Service) UpdateCombination(ctx context.Context, id string, in CreateCombinationInput) (model.LoadCombination, error) {
	kind := model.CombinationKind(in.Kind)
	if !kind.Valid() {
		return model.LoadCombination{}, fmt.Errorf("%w: kind must be strength/service", model.ErrInvalid)
	}
	combo := model.LoadCombination{
		ID:        id,
		Name:      in.Name,
		Kind:      kind,
		CoeffD:    in.CoeffD,
		CoeffL:    in.CoeffL,
		CoeffLr:   in.CoeffLr,
		CoeffS:    in.CoeffS,
		CoeffW:    in.CoeffW,
		CoeffR:    in.CoeffR,
		CoeffE:    in.CoeffE,
	}
	if !combo.HasAnyNonZero() {
		return model.LoadCombination{}, fmt.Errorf("%w: combination must have at least one non-zero coefficient", model.ErrInvalid)
	}
	// Look up project to keep event logging consistent.
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		row := tx.QueryRowContext(ctx, `SELECT project_id FROM load_combinations WHERE id=?`, id)
		return row.Scan(&p.ID)
	}); err != nil {
		return model.LoadCombination{}, wrapErr(model.ErrNotFound, "get combination project")
	}
	if err := svc.store.UpdateCombination(ctx, combo); err != nil {
		return model.LoadCombination{}, wrapErr(err, "update combination")
	}
	combo.ProjectID = p.ID
	if err := svc.appendEvent(ctx, combo.ProjectID, "update_combination", idlib.MustJSON(combo)); err != nil {
		return model.LoadCombination{}, wrapErr(err, "log update combination")
	}
	return combo, nil
}

// ListCombinations returns a project's combinations.
func (svc *Service) ListCombinations(ctx context.Context, projectID string) ([]model.LoadCombination, error) {
	return svc.store.ListCombinationsByProject(ctx, projectID)
}

// effectiveMagnitudeForKind resolves the magnitude used in a combination for a kind.
// L/S/W use the derived case if present; D/Lr/R/E use the manual case.
func effectiveMagnitudeForKind(kinds map[model.LoadKind]model.LoadCase, k model.LoadKind) (model.LoadCase, bool) {
	lc, ok := kinds[k]
	return lc, ok
}

// computeCheck computes one check row for a component under a combination.
func (svc *Service) computeCheck(ctx context.Context, tx store.DBTX, p model.Project, c model.Component, combo model.LoadCombination, eventID string) (model.Check, error) {
	cases, err := svc.allCasesForComponentInTx(ctx, tx, c.ID)
	if err != nil {
		return model.Check{}, err
	}
	// Build a per-kind effective map: prefer derived for L/S/W, else manual.
	byKind := map[model.LoadKind]model.LoadCase{}
	for _, lc := range cases {
		cur, exists := byKind[lc.Kind]
		if !exists {
			byKind[lc.Kind] = lc
			continue
		}
		// Prefer derived for L/S/W.
		if lc.Origin == model.OriginDerived && (lc.Kind == model.LoadSnow || lc.Kind == model.LoadWind) {
			byKind[lc.Kind] = lc
			continue
		}
		// Else keep existing unless existing is also same; keep manual as-is.
		if cur.Origin == model.OriginManual && lc.Origin == model.OriginManual {
			byKind[lc.Kind] = lc // last wins, arbitrary
		}
	}

	var d loadcode.DemandEffect
	addKind := func(k model.LoadKind, coef int64) {
		if coef == 0 {
			return
		}
		lc, ok := effectiveMagnitudeForKind(byKind, k)
		if !ok || lc.Magnitude == 0 {
			return
		}
		d.AddComponentDemand(c, loadcode.EffectiveMagnitude{
			Kind:      k,
			LoadType:  lc.LoadType,
			Magnitude: lc.Magnitude,
			Direction: lc.Direction,
			CoefCenti: coef,
		})
	}
	addKind(model.LoadDead, combo.CoeffD)
	addKind(model.LoadLive, combo.CoeffL)
	addKind(model.LoadRoof, combo.CoeffLr)
	addKind(model.LoadSnow, combo.CoeffS)
	addKind(model.LoadWind, combo.CoeffW)
	addKind(model.LoadRain, combo.CoeffR)
	addKind(model.LoadSeismic, combo.CoeffE)

	capM, capV, capP := loadcode.Capacities(c)
	ur := loadcode.ComputeUR(c, d, capM, capV, capP)
	return model.Check{
		ID:             idlib.NewID("chk"),
		ProjectID:      c.ProjectID,
		CombinationID:  combo.ID,
		ComponentID:    c.ID,
		DemandMoment:   d.Moment,
		DemandShear:    d.Shear,
		DemandAxial:    d.Axial,
		CapacityMoment: capM,
		CapacityShear:  capV,
		CapacityAxial:  capP,
		UR:             ur.UR,
		Status:         loadcode.StatusForUR(ur.UR),
		GoverningKind:  ur.GoverningKind,
		ComputedAt:     svc.now(),
	}, nil
}

// allCasesForComponentInTx reads every load case for a component within a transaction.
func (svc *Service) allCasesForComponentInTx(ctx context.Context, tx store.DBTX, componentID string) ([]model.LoadCase, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at
		FROM load_cases WHERE component_id=? ORDER BY CASE origin WHEN 'manual' THEN 0 ELSE 1 END, created_at DESC, id DESC`, componentID)
	if err != nil {
		return nil, fmt.Errorf("query load_cases: %w", err)
	}
	defer rows.Close()
	var out []model.LoadCase
	for rows.Next() {
		var lc model.LoadCase
		var ms int64
		var kind, lt, dir, orig string
		if err := rows.Scan(&lc.ID, &lc.ProjectID, &lc.ComponentID, &kind, &lt, &lc.Magnitude, &dir, &orig, &lc.SourceEventID, &lc.Note, &ms); err != nil {
			return nil, err
		}
		lc.Kind = model.LoadKind(kind)
		lc.LoadType = model.LoadType(lt)
		lc.Direction = model.Direction(dir)
		lc.Origin = model.Origin(orig)
		lc.CreatedAt = fromMS(ms)
		out = append(out, lc)
	}
	return out, rows.Err()
}
