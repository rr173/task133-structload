package service

import (
	"context"
	"fmt"

	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// CreateComponentInput is the validated request body for adding a structural member.
type CreateComponentInput struct {
	Code           string `json:"code"`
	Type           string `json:"type"`
	LevelID        string `json:"level_id,omitempty"`
	Span           int64  `json:"span"`            // mm
	TributaryArea  int64  `json:"tributary_area"`  // cm²
	TributaryWidth int64  `json:"tributary_width"` // mm
	KLL            int64  `json:"kll"`
	NominalMoment  int64  `json:"nominal_moment"`  // centi-kN·m
	NominalAxial   int64  `json:"nominal_axial"`   // centi-kN
	NominalShear   int64  `json:"nominal_shear"`   // centi-kN
	PhiB           int64  `json:"phi_b"`           // centi
	PhiC           int64  `json:"phi_c"`           // centi
	PhiV           int64  `json:"phi_v"`           // centi
	WindSurface    string `json:"wind_surface"`
	RoofSlope      int64  `json:"roof_slope"`      // deci-deg
	SectionLabel   string `json:"section_label"`
}

// CreateComponent validates and adds a structural member to a project.
func (svc *Service) CreateComponent(ctx context.Context, projectID string, in CreateComponentInput) (model.Component, error) {
	ct := model.ComponentType(in.Type)
	if !ct.Valid() {
		return model.Component{}, fmt.Errorf("%w: type must be beam/column/slab", model.ErrInvalid)
	}
	if in.Code == "" {
		return model.Component{}, fmt.Errorf("%w: code required", model.ErrInvalid)
	}
	if in.Span <= 0 {
		return model.Component{}, fmt.Errorf("%w: span must be > 0", model.ErrInvalid)
	}
	if in.TributaryArea < 0 {
		return model.Component{}, fmt.Errorf("%w: tributary_area must be >= 0", model.ErrInvalid)
	}
	if in.KLL <= 0 {
		return model.Component{}, fmt.Errorf("%w: kll must be > 0", model.ErrInvalid)
	}
	if in.PhiB < 0 || in.PhiB > 100 {
		return model.Component{}, fmt.Errorf("%w: phi_b must be in [0,100]", model.ErrInvalid)
	}
	ws := model.WindSurface(in.WindSurface)
	if !ws.Valid() {
		return model.Component{}, fmt.Errorf("%w: invalid wind_surface", model.ErrInvalid)
	}
	if in.RoofSlope < 0 {
		return model.Component{}, fmt.Errorf("%w: roof_slope must be >= 0", model.ErrInvalid)
	}
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		// If a level is given, resolve the project from it (level-scoped route);
		// otherwise require projectID directly.
		if projectID == "" && in.LevelID != "" {
			lvl, err := svc.store.GetLevelInTx(ctx, tx, in.LevelID)
			if err != nil {
				return err
			}
			projectID = lvl.ProjectID
		}
		var err error
		p, err = svc.store.GetProjectInTx(ctx, tx, projectID)
		if err != nil {
			return err
		}
		if in.LevelID != "" {
			if _, err := svc.store.GetLevelInTx(ctx, tx, in.LevelID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return model.Component{}, wrapErr(err, "resolve project/level")
	}
	c := model.Component{
		ID:             idlib.NewID("cmp"),
		ProjectID:      p.ID,
		LevelID:        in.LevelID,
		Code:           in.Code,
		Type:           ct,
		Span:           in.Span,
		TributaryArea:  in.TributaryArea,
		TributaryWidth: in.TributaryWidth,
		KLL:            in.KLL,
		NominalMoment:  in.NominalMoment,
		NominalAxial:   in.NominalAxial,
		NominalShear:   in.NominalShear,
		PhiB:           in.PhiB,
		PhiC:           in.PhiC,
		PhiV:           in.PhiV,
		WindSurface:    ws,
		RoofSlope:      in.RoofSlope,
		SectionLabel:   in.SectionLabel,
		CreatedAt:      svc.now(),
	}
	if err := svc.store.CreateComponent(ctx, c); err != nil {
		return model.Component{}, wrapErr(err, "create component")
	}
	if err := svc.appendEvent(ctx, c.ProjectID, "add_component", idlib.MustJSON(c)); err != nil {
		return model.Component{}, wrapErr(err, "log add component")
	}
	return c, nil
}

// UpdateComponentInput edits authoritative component fields (does not touch derived state).
type UpdateComponentInput struct {
	Code           *string `json:"code,omitempty"`
	Type           *string `json:"type,omitempty"`
	Span           *int64  `json:"span,omitempty"`
	TributaryArea  *int64  `json:"tributary_area,omitempty"`
	TributaryWidth *int64  `json:"tributary_width,omitempty"`
	KLL            *int64  `json:"kll,omitempty"`
	NominalMoment  *int64  `json:"nominal_moment,omitempty"`
	NominalAxial   *int64  `json:"nominal_axial,omitempty"`
	NominalShear   *int64  `json:"nominal_shear,omitempty"`
	PhiB           *int64  `json:"phi_b,omitempty"`
	PhiC           *int64  `json:"phi_c,omitempty"`
	PhiV           *int64  `json:"phi_v,omitempty"`
	WindSurface    *string `json:"wind_surface,omitempty"`
	RoofSlope      *int64  `json:"roof_slope,omitempty"`
	SectionLabel   *string `json:"section_label,omitempty"`
}

// UpdateComponent applies authoritative field edits.
func (svc *Service) UpdateComponent(ctx context.Context, id string, in UpdateComponentInput) (model.Component, error) {
	var c model.Component
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		c, err = svc.store.GetComponentInTx(ctx, tx, id)
		return err
	}); err != nil {
		return model.Component{}, wrapErr(err, "get component")
	}
	if in.Code != nil {
		c.Code = *in.Code
	}
	if in.Type != nil {
		ct := model.ComponentType(*in.Type)
		if !ct.Valid() {
			return model.Component{}, fmt.Errorf("%w: type must be beam/column/slab", model.ErrInvalid)
		}
		c.Type = ct
	}
	if in.Span != nil {
		if *in.Span <= 0 {
			return model.Component{}, fmt.Errorf("%w: span must be > 0", model.ErrInvalid)
		}
		c.Span = *in.Span
	}
	if in.TributaryArea != nil {
		if *in.TributaryArea < 0 {
			return model.Component{}, fmt.Errorf("%w: tributary_area must be >= 0", model.ErrInvalid)
		}
		c.TributaryArea = *in.TributaryArea
	}
	if in.TributaryWidth != nil {
		c.TributaryWidth = *in.TributaryWidth
	}
	if in.KLL != nil {
		if *in.KLL <= 0 {
			return model.Component{}, fmt.Errorf("%w: kll must be > 0", model.ErrInvalid)
		}
		c.KLL = *in.KLL
	}
	if in.NominalMoment != nil {
		c.NominalMoment = *in.NominalMoment
	}
	if in.NominalAxial != nil {
		c.NominalAxial = *in.NominalAxial
	}
	if in.NominalShear != nil {
		c.NominalShear = *in.NominalShear
	}
	if in.PhiB != nil {
		if *in.PhiB < 0 || *in.PhiB > 100 {
			return model.Component{}, fmt.Errorf("%w: phi_b must be in [0,100]", model.ErrInvalid)
		}
		c.PhiB = *in.PhiB
	}
	if in.PhiC != nil {
		c.PhiC = *in.PhiC
	}
	if in.PhiV != nil {
		c.PhiV = *in.PhiV
	}
	if in.WindSurface != nil {
		ws := model.WindSurface(*in.WindSurface)
		if !ws.Valid() {
			return model.Component{}, fmt.Errorf("%w: invalid wind_surface", model.ErrInvalid)
		}
		c.WindSurface = ws
	}
	if in.RoofSlope != nil {
		if *in.RoofSlope < 0 {
			return model.Component{}, fmt.Errorf("%w: roof_slope must be >= 0", model.ErrInvalid)
		}
		c.RoofSlope = *in.RoofSlope
	}
	if in.SectionLabel != nil {
		c.SectionLabel = *in.SectionLabel
	}
	if err := svc.store.UpdateComponent(ctx, c); err != nil {
		return model.Component{}, wrapErr(err, "update component")
	}
	if err := svc.appendEvent(ctx, c.ProjectID, "update_component", idlib.MustJSON(c)); err != nil {
		return model.Component{}, wrapErr(err, "log update component")
	}
	return c, nil
}

// GetComponent fetches a member.
func (svc *Service) GetComponent(ctx context.Context, id string) (model.Component, error) {
	var c model.Component
	var err error
	svc.store.InTx(ctx, func(tx store.DBTX) error {
		c, err = svc.store.GetComponentInTx(ctx, tx, id)
		return nil
	})
	return c, err
}

// ListComponentsByLevel and by project delegate to store.
func (svc *Service) ListComponentsByLevel(ctx context.Context, levelID string) ([]model.Component, error) {
	return svc.store.ListComponentsByLevel(ctx, levelID)
}

func (svc *Service) ListComponentsByProject(ctx context.Context, projectID string) ([]model.Component, error) {
	return svc.store.ListComponentsByProject(ctx, projectID)
}
