package store

import (
	"context"
	"fmt"
	"strconv"

	"task133-structload/internal/model"
)

// CreateComponent inserts a structural member.
func (s *Store) CreateComponent(ctx context.Context, c model.Component) error {
	const cols = `(id,project_id,level_id,code,type,span,tributary_area,tributary_width,kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at)`
	const ph = `(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.ExecContext(ctx, `INSERT INTO components `+cols+` VALUES`+ph,
		c.ID, c.ProjectID, c.LevelID, c.Code, string(c.Type), c.Span, c.TributaryArea, c.TributaryWidth,
		c.KLL, c.NominalMoment, c.NominalAxial, c.NominalShear, c.PhiB, c.PhiC, c.PhiV,
		string(c.WindSurface), c.RoofSlope, c.SectionLabel, nowMS(c.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert component: %w", err)
	}
	return nil
}

// UpdateComponent writes authoritative fields. Derived state is unaffected.
func (s *Store) UpdateComponent(ctx context.Context, c model.Component) error {
	res, err := s.db.ExecContext(ctx, `UPDATE components SET code=?,type=?,span=?,tributary_area=?,tributary_width=?,
		kll=?,nominal_moment=?,nominal_axial=?,nominal_shear=?,phi_b=?,phi_c=?,phi_v=?,wind_surface=?,roof_slope=?,section_label=? WHERE id=?`,
		c.Code, string(c.Type), c.TributaryWidth, c.TributaryArea, c.TributaryWidth,
		c.KLL, c.NominalMoment, c.NominalAxial, c.NominalShear, c.PhiB, c.PhiC, c.PhiV,
		string(c.WindSurface), c.RoofSlope, c.SectionLabel, c.ID)
	if err != nil {
		return fmt.Errorf("update component: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// GetComponentInTx fetches a component inside a transaction.
func (s *Store) GetComponentInTx(ctx context.Context, tx DBTX, id string) (model.Component, error) {
	var c model.Component
	var ms int64
	var t, ws string
	err := tx.QueryRowContext(ctx, `SELECT id,project_id,level_id,code,type,span,tributary_area,tributary_width,
		kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at
		FROM components WHERE id=?`, id).Scan(
		&c.ID, &c.ProjectID, &c.LevelID, &c.Code, &t, &c.Span, &c.TributaryArea, &c.TributaryWidth,
		&c.KLL, &c.NominalMoment, &c.NominalAxial, &c.NominalShear, &c.PhiB, &c.PhiC, &c.PhiV,
		&ws, &c.RoofSlope, &c.SectionLabel, &ms)
	if err != nil {
		return c, model.ErrNotFound
	}
	c.Type = model.ComponentType(t)
	c.WindSurface = model.WindSurface(ws)
	c.CreatedAt = fromMS(ms)
	return c, nil
}

// ListComponentsByLevel returns components of a level.
func (s *Store) ListComponentsByLevel(ctx context.Context, levelID string) ([]model.Component, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,level_id,code,type,span,tributary_area,tributary_width,
		kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at
		FROM components WHERE level_id=? ORDER BY created_at`, levelID)
	if err != nil {
		return nil, fmt.Errorf("list components by level: %w", err)
	}
	defer rows.Close()
	return scanComponents(rows)
}

// ListComponentsByProject returns components of a project.
func (s *Store) ListComponentsByProject(ctx context.Context, projectID string) ([]model.Component, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,level_id,code,type,span,tributary_area,tributary_width,
		kll,nominal_moment,nominal_axial,nominal_shear,phi_b,phi_c,phi_v,wind_surface,roof_slope,section_label,created_at
		FROM components WHERE project_id=? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list components by project: %w", err)
	}
	defer rows.Close()
	return scanComponents(rows)
}

// scanComponents materializes component rows from a *sql.Rows.
func scanComponents(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]model.Component, error) {
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

// intify helpers used by override application.
func mustParseInt(s string) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}
