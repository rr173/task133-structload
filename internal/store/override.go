package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// CreateOverride inserts a post-computation adjustment record.
func (s *Store) CreateOverride(ctx context.Context, o model.Override) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO overrides
		(id,project_id,component_id,field,old_value,new_value,reason,applied_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		o.ID, o.ProjectID, o.ComponentID, o.Field, o.OldValue, o.NewValue, o.Reason, nowMS(o.AppliedAt))
	if err != nil {
		return fmt.Errorf("insert override: %w", err)
	}
	return nil
}

// ListOverridesByProject returns overrides ordered by application time.
func (s *Store) ListOverridesByProject(ctx context.Context, projectID string) ([]model.Override, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,component_id,field,old_value,new_value,reason,applied_at
		FROM overrides WHERE project_id=? ORDER BY applied_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list overrides: %w", err)
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
