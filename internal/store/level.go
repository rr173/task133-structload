package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// CreateLevel inserts a floor level.
func (s *Store) CreateLevel(ctx context.Context, l model.Level) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO levels (id,project_id,name,elevation,gross_area,created_at)
		VALUES(?,?,?,?,?,?)`, l.ID, l.ProjectID, l.Name, l.Elevation, l.GrossArea, nowMS(l.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert level: %w", err)
	}
	return nil
}

// ListLevelsByProject returns levels of a project ordered by elevation then creation.
func (s *Store) ListLevelsByProject(ctx context.Context, projectID string) ([]model.Level, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,name,elevation,gross_area,created_at
		FROM levels WHERE project_id=? ORDER BY elevation, created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list levels: %w", err)
	}
	defer rows.Close()
	var out []model.Level
	for rows.Next() {
		var l model.Level
		var ms int64
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.Name, &l.Elevation, &l.GrossArea, &ms); err != nil {
			return nil, err
		}
		l.CreatedAt = fromMS(ms)
		out = append(out, l)
	}
	return out, rows.Err()
}

// GetLevelInTx fetches a level inside a transaction.
func (s *Store) GetLevelInTx(ctx context.Context, tx DBTX, id string) (model.Level, error) {
	var l model.Level
	var ms int64
	err := tx.QueryRowContext(ctx, `SELECT id,project_id,name,elevation,gross_area,created_at FROM levels WHERE id=?`, id).
		Scan(&l.ID, &l.ProjectID, &l.Name, &l.Elevation, &l.GrossArea, &ms)
	if err != nil {
		return l, model.ErrNotFound
	}
	l.CreatedAt = fromMS(ms)
	return l, nil
}
