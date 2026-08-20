package service

import (
	"context"
	"fmt"
	"time"

	"task133-structload/internal/idlib"
	"task133-structload/internal/loadcode"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// AddManualLoadCaseInput records an authoritative load case (D/L/Lr/R/E).
type AddManualLoadCaseInput struct {
	Kind      string `json:"kind"`
	LoadType  string `json:"load_type"`
	Magnitude int64  `json:"magnitude"`
	Direction string `json:"direction"`
	Note      string `json:"note,omitempty"`
}

// AddManualLoadCase validates and stores a manual load case.
func (svc *Service) AddManualLoadCase(ctx context.Context, componentID string, in AddManualLoadCaseInput) (model.LoadCase, error) {
	kind := model.LoadKind(in.Kind)
	if !kind.Valid() || kind == model.LoadSnow || kind == model.LoadWind {
		return model.LoadCase{}, fmt.Errorf("%w: kind must be D/L/Lr/R/E", model.ErrInvalid)
	}
	lt := model.LoadType(in.LoadType)
	if !lt.Valid() {
		return model.LoadCase{}, fmt.Errorf("%w: load_type must be area_pressure/line_load/point_load", model.ErrInvalid)
	}
	dir := model.Direction(in.Direction)
	if !dir.Valid() {
		return model.LoadCase{}, fmt.Errorf("%w: direction must be gravity/uplift/lateral", model.ErrInvalid)
	}
	if in.Magnitude < 0 {
		return model.LoadCase{}, fmt.Errorf("%w: magnitude must be >= 0", model.ErrInvalid)
	}
	var c model.Component
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		c, err = svc.store.GetComponentInTx(ctx, tx, componentID)
		return err
	}); err != nil {
		return model.LoadCase{}, wrapErr(err, "get component")
	}
	lc := model.LoadCase{
		ID:          idlib.NewID("lc"),
		ProjectID:   c.ProjectID,
		ComponentID: c.ID,
		Kind:        kind,
		LoadType:    lt,
		Magnitude:   in.Magnitude,
		Direction:   model.DirGravity,
		Origin:      model.OriginManual,
		Note:        in.Note,
		CreatedAt:   svc.now(),
	}
	if err := svc.store.CreateLoadCase(ctx, lc); err != nil {
		return model.LoadCase{}, wrapErr(err, "create load_case")
	}
	if err := svc.appendEvent(ctx, lc.ProjectID, "add_loadcase", idlib.MustJSON(lc)); err != nil {
		return model.LoadCase{}, wrapErr(err, "log add loadcase")
	}
	return lc, nil
}

// ListLoadCases returns all load cases for a component.
func (svc *Service) ListLoadCases(ctx context.Context, componentID string) ([]model.LoadCase, error) {
	return svc.store.ListLoadCasesByComponent(ctx, componentID)
}

// ListDerivedCases returns only derived cases for a component.
func (svc *Service) ListDerivedCases(ctx context.Context, componentID string) ([]model.LoadCase, error) {
	return svc.store.ListDerivedCasesByComponent(ctx, componentID)
}

// deriveForComponent computes the W/S/reduced-L derived load cases for one component
// against project wind/snow parameters. Returns derived cases (not yet persisted).
func (svc *Service) deriveForComponent(ctx context.Context, tx store.DBTX, p model.Project, c model.Component, eventID string) ([]model.LoadCase, error) {
	var out []model.LoadCase
	now := svc.now()

	// Wind pressure W (only if component has a wind surface).
	if c.WindSurface != model.WindNone && c.WindSurface != "" {
		heightMM := int64(0)
		if lvl, err := svc.store.GetLevelInTx(ctx, tx, c.LevelID); err == nil {
			heightMM = lvl.Elevation
		}
		if heightMM == 0 {
			heightMM = 10000 // default 10 m
		}
		qz := loadcode.VelocityPressureCenti(string(p.Exposure), heightMM, p.Kzt, p.Kd, p.WindSpeed)
		cpExt := c.WindSurface.CpExternal()
		pWind := loadcode.WindPressureCenti(qz, cpExt, p.GustG)
		if pWind != 0 {
			out = append(out, model.LoadCase{
				ID:            idlib.NewID("lc"),
				ProjectID:     p.ID,
				ComponentID:   c.ID,
				Kind:          model.LoadWind,
				LoadType:      model.LoadAreaPressure,
				Magnitude:     pWind,
				Direction:     c.WindSurface.DirectionForWind(),
				Origin:        model.OriginDerived,
				SourceEventID: eventID,
				Note:          fmt.Sprintf("qz=%d cp=%g", qz, cpExt),
				CreatedAt:     now,
			})
		}
	}

	// Snow load S (only if there is ground snow).
	if p.GroundSnow > 0 {
		pf, ps := loadcode.SnowLoadCenti(p.SnowExposureFactor, p.ThermalFactor,
			model.ImportanceFactorIs(p.Importance), p.GroundSnow, c.RoofSlope)
		snow := ps
		if c.RoofSlope == 0 {
			snow = pf // flat roof
		}
		if snow != 0 {
			out = append(out, model.LoadCase{
				ID:            idlib.NewID("lc"),
				ProjectID:     p.ID,
				ComponentID:   c.ID,
				Kind:          model.LoadSnow,
				LoadType:      model.LoadAreaPressure,
				Magnitude:     snow,
				Direction:     model.DirGravity,
				Origin:        model.OriginDerived,
				SourceEventID: eventID,
				Note:          fmt.Sprintf("pf=%d ps=%d", pf, ps),
				CreatedAt:     now,
			})
		}
	}

	// Reduced live load L (only if there is a manual L=Lo case).
	manuals, err := svc.manualCasesInTx(ctx, tx, c.ID)
	if err != nil {
		return nil, err
	}
	for _, mc := range manuals {
		if mc.Kind != model.LoadLive {
			continue
		}
		reduced := loadcode.ReducedLiveLoadCenti(c.KLL, c.TributaryArea, mc.Magnitude)
		out = append(out, model.LoadCase{
			ID:            idlib.NewID("lc"),
			ProjectID:     p.ID,
			ComponentID:   c.ID,
			Kind:          model.LoadLive,
			LoadType:      mc.LoadType,
			Magnitude:     reduced,
			Direction:     mc.Direction,
			Origin:        model.OriginDerived,
			SourceEventID: eventID,
			Note:          fmt.Sprintf("reduced Lo=%d -> L=%d", mc.Magnitude, reduced),
			CreatedAt:     now,
		})
		break // one reduced L per component
	}
	return out, nil
}

// manualCasesInTx reads manual load cases for a component within a transaction.
func (svc *Service) manualCasesInTx(ctx context.Context, tx store.DBTX, componentID string) ([]model.LoadCase, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at
		FROM load_cases WHERE component_id=? AND origin='manual' ORDER BY kind`, componentID)
	if err != nil {
		return nil, fmt.Errorf("query manual load_cases: %w", err)
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

// fromMS mirrors store.fromMS; local copy to avoid a package cycle in tx helpers.
func fromMS(ms int64) time.Time { return time.UnixMilli(ms).UTC() }
