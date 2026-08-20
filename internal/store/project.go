package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// CreateProject inserts a project.
func (s *Store) CreateProject(ctx context.Context, p model.Project) error {
	const cols = `(id,code,name,site,wind_zone,snow_zone,exposure,importance,wind_speed,kzt,kd,gust_g,ground_snow,snow_exposure_factor,thermal_factor,units,created_at)`
	const ph = `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.ExecContext(ctx, `INSERT INTO projects `+cols+` VALUES`+ph,
		p.ID, p.Code, p.Name, p.Site, p.WindZone, p.SnowZone, string(p.Exposure), string(p.Importance),
		p.WindSpeed, p.Kzt, p.Kd, p.GustG, p.GroundSnow, p.SnowExposureFactor, p.ThermalFactor, p.Units, nowMS(p.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}
	return nil
}

// UpdateProject writes wind/snow params for an existing project.
func (s *Store) UpdateProject(ctx context.Context, p model.Project) error {
	res, err := s.db.ExecContext(ctx, `UPDATE projects SET code=?,name=?,site=?,wind_zone=?,snow_zone=?,exposure=?,importance=?,
		wind_speed=?,kzt=?,kd=?,gust_g=?,ground_snow=?,snow_exposure_factor=?,thermal_factor=?,units=? WHERE id=?`,
		p.Code, p.Name, p.Site, p.WindZone, p.SnowZone, string(p.Exposure), string(p.Importance),
		p.WindSpeed, p.Kzt, p.Kd, p.GustG, p.GroundSnow, p.SnowExposureFactor, p.ThermalFactor, p.Units, p.ID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// DeleteProject removes a project and its dependents (cascade).
func (s *Store) DeleteProject(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// GetProjectInTx fetches a project within a transaction.
func (s *Store) GetProjectInTx(ctx context.Context, tx DBTX, id string) (model.Project, error) {
	var p model.Project
	var ms int64
	err := tx.QueryRowContext(ctx, `SELECT id,code,name,site,wind_zone,snow_zone,exposure,importance,
		wind_speed,kzt,kd,gust_g,ground_snow,snow_exposure_factor,thermal_factor,units,created_at
		FROM projects LIMIT 1`).Scan(
		&p.ID, &p.Code, &p.Name, &p.Site, &p.WindZone, &p.SnowZone, &p.Exposure, &p.Importance,
		&p.WindSpeed, &p.Kzt, &p.Kd, &p.GustG, &p.GroundSnow, &p.SnowExposureFactor, &p.ThermalFactor, &p.Units, &ms)
	if err != nil {
		return p, model.ErrNotFound
	}
	p.CreatedAt = fromMS(ms)
	return p, nil
}

// ListProjects returns all projects ordered by creation.
func (s *Store) ListProjects(ctx context.Context) ([]model.Project, error) {
	snap, err := s.LoadAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.Project, 0, len(snap.Projects))
	for _, p := range snap.Projects {
		out = append(out, p)
	}
	return out, nil
}
