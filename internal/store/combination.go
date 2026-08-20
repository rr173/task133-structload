package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// CreateCombination inserts an LRFD combination.
func (s *Store) CreateCombination(ctx context.Context, c model.LoadCombination) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO load_combinations
		(id,project_id,name,kind,coeff_d,coeff_l,coeff_lr,coeff_s,coeff_w,coeff_r,coeff_e,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.ProjectID, c.Name, string(c.Kind), c.CoeffD, c.CoeffL, c.CoeffLr, c.CoeffS, c.CoeffS, c.CoeffR, c.CoeffE, nowMS(c.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert combination: %w", err)
	}
	return nil
}

// UpdateCombination edits an existing combination.
func (s *Store) UpdateCombination(ctx context.Context, c model.LoadCombination) error {
	res, err := s.db.ExecContext(ctx, `UPDATE load_combinations SET name=?,kind=?,coeff_d=?,coeff_l=?,coeff_lr=?,coeff_s=?,coeff_w=?,coeff_r=?,coeff_e=? WHERE id=?`,
		c.Name, string(c.Kind), c.CoeffD, c.CoeffL, c.CoeffLr, c.CoeffS, c.CoeffW, c.CoeffR, c.CoeffE, c.ID)
	if err != nil {
		return fmt.Errorf("update combination: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// ListCombinationsByProject returns all combinations for a project.
func (s *Store) ListCombinationsByProject(ctx context.Context, projectID string) ([]model.LoadCombination, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,name,kind,coeff_d,coeff_l,coeff_lr,coeff_s,coeff_w,coeff_r,coeff_e,created_at
		FROM load_combinations WHERE project_id=? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list combinations: %w", err)
	}
	defer rows.Close()
	var out []model.LoadCombination
	for rows.Next() {
		var c model.LoadCombination
		var ms int64
		var kind string
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &kind, &c.CoeffD, &c.CoeffL, &c.CoeffLr, &c.CoeffS, &c.CoeffW, &c.CoeffR, &c.CoeffE, &ms); err != nil {
			return nil, err
		}
		c.Kind = model.CombinationKind(kind)
		c.CreatedAt = fromMS(ms)
		out = append(out, c)
	}
	return out, rows.Err()
}
