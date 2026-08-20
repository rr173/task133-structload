package service

import (
	"context"
	"fmt"

	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// CreateProjectInput is the validated request body for creating a project.
type CreateProjectInput struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	Site          string `json:"site"`
	WindZone      string `json:"wind_zone"`
	SnowZone      string `json:"snow_zone"`
	Exposure      string `json:"exposure"`
	Importance    string `json:"importance"`
	WindSpeed   int64  `json:"wind_speed"`     // deci-m/s
	Kzt         *int64 `json:"kzt,omitempty"`  // centi; default 100
	Kd          *int64 `json:"kd,omitempty"`   // centi; default 85
	GustG       *int64 `json:"gust_g,omitempty"` // centi; default 85
	GroundSnow  int64  `json:"ground_snow"`    // centi-kPa
	SnowExposure *int64 `json:"snow_exposure_factor,omitempty"` // centi; default 100
	ThermalFactor *int64 `json:"thermal_factor,omitempty"`      // centi; default 100
	Units         string `json:"units"`
}

// CreateProject validates and persists a new project.
func (svc *Service) CreateProject(ctx context.Context, in CreateProjectInput) (model.Project, error) {
	if in.Code == "" || in.Name == "" {
		return model.Project{}, fmt.Errorf("%w: code and name required", model.ErrInvalid)
	}
	exp := model.ExposureCategory(in.Exposure)
	if !exp.Valid() {
		return model.Project{}, fmt.Errorf("%w: exposure must be B/C/D", model.ErrInvalid)
	}
	imp := model.ImportanceCategory(in.Importance)
	if !imp.Valid() {
		return model.Project{}, fmt.Errorf("%w: importance must be I/II/III/IV", model.ErrInvalid)
	}
	if in.WindSpeed <= 0 {
		return model.Project{}, fmt.Errorf("%w: wind_speed must be > 0", model.ErrInvalid)
	}
	if in.GroundSnow < 0 {
		return model.Project{}, fmt.Errorf("%w: ground_snow must be >= 0", model.ErrInvalid)
	}
	kzt := ptrOr(in.Kzt, 100)
	kd := ptrOr(in.Kd, 85)
	gust := ptrOr(in.GustG, 85)
	ce := ptrOr(in.SnowExposure, 100)
	ct := ptrOr(in.ThermalFactor, 100)
	if in.Units == "" {
		in.Units = "SI"
	}
	p := model.Project{
		ID:                 idlib.NewID("prj"),
		Code:               in.Code,
		Name:               in.Name,
		Site:               in.Site,
		WindZone:           in.WindZone,
		SnowZone:           in.SnowZone,
		Exposure:           exp,
		Importance:         imp,
		WindSpeed:          in.WindSpeed,
		Kzt:                kzt,
		Kd:                 kd,
		GustG:              gust,
		GroundSnow:         in.GroundSnow,
		SnowExposureFactor: ce,
		ThermalFactor:      ct,
		Units:              in.Units,
		CreatedAt:          svc.now(),
	}
	if err := svc.store.CreateProject(ctx, p); err != nil {
		return model.Project{}, wrapErr(err, "create project")
	}
	if err := svc.appendEvent(ctx, p.ID, "create_project", idlib.MustJSON(p)); err != nil {
		return model.Project{}, wrapErr(err, "log create project")
	}
	return p, nil
}

// UpdateProjectInput edits wind/snow params.
type UpdateProjectInput struct {
	Code          *string `json:"code,omitempty"`
	Name          *string `json:"name,omitempty"`
	Site          *string `json:"site,omitempty"`
	WindZone      *string `json:"wind_zone,omitempty"`
	SnowZone      *string `json:"snow_zone,omitempty"`
	Exposure      *string `json:"exposure,omitempty"`
	Importance    *string `json:"importance,omitempty"`
	WindSpeed     *int64  `json:"wind_speed,omitempty"`
	Kzt           *int64  `json:"kzt,omitempty"`
	Kd            *int64  `json:"kd,omitempty"`
	GustG         *int64  `json:"gust_g,omitempty"`
	GroundSnow    *int64  `json:"ground_snow,omitempty"`
	SnowExposure  *int64  `json:"snow_exposure_factor,omitempty"`
	ThermalFactor *int64  `json:"thermal_factor,omitempty"`
	Units         *string `json:"units,omitempty"`
}

// UpdateProject applies field-level updates to a project.
func (svc *Service) UpdateProject(ctx context.Context, id string, in UpdateProjectInput) (model.Project, error) {
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		p, err = svc.store.GetProjectInTx(ctx, tx, id)
		return err
	}); err != nil {
		return model.Project{}, wrapErr(err, "get project")
	}
	if in.Code != nil {
		p.Code = *in.Code
	}
	if in.Name != nil {
		p.Name = *in.Name
	}
	if in.Site != nil {
		p.Site = *in.Site
	}
	if in.WindZone != nil {
		p.WindZone = *in.WindZone
	}
	if in.SnowZone != nil {
		p.SnowZone = *in.SnowZone
	}
	if in.Exposure != nil {
		exp := model.ExposureCategory(*in.Exposure)
		if !exp.Valid() {
			return model.Project{}, fmt.Errorf("%w: exposure must be B/C/D", model.ErrInvalid)
		}
		p.Exposure = exp
	}
	if in.Importance != nil {
		imp := model.ImportanceCategory(*in.Importance)
		if !imp.Valid() {
			return model.Project{}, fmt.Errorf("%w: importance must be I/II/III/IV", model.ErrInvalid)
		}
		p.Importance = imp
	}
	if in.WindSpeed != nil {
		if *in.WindSpeed <= 0 {
			return model.Project{}, fmt.Errorf("%w: wind_speed must be > 0", model.ErrInvalid)
		}
		p.WindSpeed = p.WindSpeed
	}
	if in.Kzt != nil {
		p.Kzt = *in.Kzt
	}
	if in.Kd != nil {
		p.Kd = *in.Kd
	}
	if in.GustG != nil {
		p.GustG = *in.GustG
	}
	if in.GroundSnow != nil {
		if *in.GroundSnow < 0 {
			return model.Project{}, fmt.Errorf("%w: ground_snow must be >= 0", model.ErrInvalid)
		}
		p.GroundSnow = *in.GroundSnow
	}
	if in.SnowExposure != nil {
		p.SnowExposureFactor = *in.SnowExposure
	}
	if in.ThermalFactor != nil {
		p.ThermalFactor = *in.ThermalFactor
	}
	if in.Units != nil {
		p.Units = *in.Units
	}
	if err := svc.store.UpdateProject(ctx, p); err != nil {
		return model.Project{}, wrapErr(err, "update project")
	}
	if err := svc.appendEvent(ctx, p.ID, "update_project", idlib.MustJSON(p)); err != nil {
		return model.Project{}, wrapErr(err, "log update project")
	}
	return p, nil
}

// GetProject fetches a project by id.
func (svc *Service) GetProject(ctx context.Context, id string) (model.Project, error) {
	var p model.Project
	var err error
	svc.store.InTx(ctx, func(tx store.DBTX) error {
		p, err = svc.store.GetProjectInTx(ctx, tx, id)
		return nil
	})
	return p, err
}

// ListProjects returns all projects.
func (svc *Service) ListProjects(ctx context.Context) ([]model.Project, error) {
	return svc.store.ListProjects(ctx)
}

// DeleteProject removes a project (cascade).
func (svc *Service) DeleteProject(ctx context.Context, id string) error {
	if err := svc.store.DeleteProject(ctx, id); err != nil {
		return wrapErr(err, "delete project")
	}
	return nil
}

// CreateLevelInput adds a floor.
type CreateLevelInput struct {
	Name      string `json:"name"`
	Elevation int64  `json:"elevation"`  // mm
	GrossArea int64  `json:"gross_area"`  // cm²
}

// CreateLevel adds a floor to a project.
func (svc *Service) CreateLevel(ctx context.Context, projectID string, in CreateLevelInput) (model.Level, error) {
	if in.Name == "" {
		return model.Level{}, fmt.Errorf("%w: name required", model.ErrInvalid)
	}
	if in.GrossArea < 0 {
		return model.Level{}, fmt.Errorf("%w: gross_area must be >= 0", model.ErrInvalid)
	}
	var p model.Project
	if err := svc.store.InTx(ctx, func(tx store.DBTX) error {
		var err error
		p, err = svc.store.GetProjectInTx(ctx, tx, projectID)
		return err
	}); err != nil {
		return model.Level{}, wrapErr(err, "get project")
	}
	l := model.Level{
		ID:        idlib.NewID("lvl"),
		ProjectID: p.ID,
		Name:      in.Name,
		Elevation: in.Elevation,
		GrossArea: in.GrossArea,
		CreatedAt: svc.now(),
	}
	if err := svc.store.CreateLevel(ctx, l); err != nil {
		return model.Level{}, wrapErr(err, "create level")
	}
	if err := svc.appendEvent(ctx, l.ProjectID, "add_level", idlib.MustJSON(l)); err != nil {
		return model.Level{}, wrapErr(err, "log add level")
	}
	return l, nil
}

// ListLevels returns a project's floors.
func (svc *Service) ListLevels(ctx context.Context, projectID string) ([]model.Level, error) {
	return svc.store.ListLevelsByProject(ctx, projectID)
}

func ptrOr(p *int64, def int64) int64 {
	if p != nil {
		return *p
	}
	return def
}
