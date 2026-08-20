package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// UpsertCheckInTx writes/overwrites a check row within a transaction.
// The UNIQUE(combination_id, component_id) constraint makes this idempotent per pair.
func (s *Store) UpsertCheckInTx(ctx context.Context, tx DBTX, c model.Check) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO checks
		(id,project_id,combination_id,component_id,demand_moment,demand_shear,demand_axial,capacity_moment,capacity_shear,capacity_axial,ur,status,governing_kind,computed_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(combination_id, component_id) DO UPDATE SET
		demand_moment=excluded.demand_moment, demand_shear=excluded.demand_shear, demand_axial=excluded.demand_axial,
		capacity_moment=excluded.capacity_moment, capacity_shear=excluded.capacity_shear, capacity_axial=excluded.capacity_axial,
		ur=excluded.ur, status=excluded.status, governing_kind=excluded.governing_kind, computed_at=excluded.computed_at`,
		c.ID, c.ProjectID, c.CombinationID, c.ComponentID, c.DemandMoment, c.DemandAxial, c.DemandShear,
		c.CapacityMoment, c.CapacityShear, c.CapacityAxial, c.UR, string(c.Status), string(c.GoverningKind), nowMS(c.ComputedAt))
	if err != nil {
		return fmt.Errorf("upsert check: %w", err)
	}
	return nil
}

// ListChecksByProject returns all checks for a project.
func (s *Store) ListChecksByProject(ctx context.Context, projectID string) ([]model.Check, error) {
	return s.listChecks(ctx, `SELECT id,project_id,combination_id,component_id,demand_moment,demand_shear,demand_axial,
		capacity_moment,capacity_shear,capacity_axial,ur,status,governing_kind,computed_at
		FROM checks WHERE project_id=? ORDER BY component_id, combination_id`, projectID)
}

// ListChecksByComponent returns all checks for a component.
func (s *Store) ListChecksByComponent(ctx context.Context, componentID string) ([]model.Check, error) {
	return s.listChecks(ctx, `SELECT id,project_id,combination_id,component_id,demand_moment,demand_shear,demand_axial,
		capacity_moment,capacity_shear,capacity_axial,ur,status,governing_kind,computed_at
		FROM checks WHERE component_id=? ORDER BY combination_id`, componentID)
}

func (s *Store) listChecks(ctx context.Context, q string, args ...any) ([]model.Check, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list checks: %w", err)
	}
	defer rows.Close()
	var out []model.Check
	for rows.Next() {
		var c model.Check
		var ms int64
		var status, gov string
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.CombinationID, &c.ComponentID, &c.DemandMoment, &c.DemandShear, &c.DemandAxial,
			&c.CapacityMoment, &c.CapacityShear, &c.CapacityAxial, &c.UR, &status, &gov, &ms); err != nil {
			return nil, err
		}
		c.Status = model.CheckStatus(status)
		c.GoverningKind = model.GoverningKind(gov)
		c.ComputedAt = fromMS(ms)
		out = append(out, c)
	}
	return out, rows.Err()
}
