package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// CreateLoadCase inserts a load case (manual or derived).
func (s *Store) CreateLoadCase(ctx context.Context, lc model.LoadCase) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO load_cases
		(id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		lc.ID, lc.ProjectID, lc.ComponentID, string(lc.Kind), string(lc.LoadType), lc.Magnitude,
		string(lc.Direction), string(lc.Origin), lc.SourceEventID, lc.Note, nowMS(lc.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert load_case: %w", err)
	}
	return nil
}

// CreateLoadCasesInTx inserts many derived load cases in one transaction.
func (s *Store) CreateLoadCasesInTx(ctx context.Context, tx DBTX, cases []model.LoadCase) error {
	for _, lc := range cases {
		if _, err := tx.ExecContext(ctx, `INSERT INTO load_cases
			(id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			lc.ID, lc.ProjectID, lc.ComponentID, string(lc.Kind), string(lc.LoadType), lc.Magnitude,
			string(lc.Direction), string(lc.Origin), lc.SourceEventID, lc.Note, nowMS(lc.CreatedAt)); err != nil {
			return fmt.Errorf("insert load_case: %w", err)
		}
	}
	return nil
}

// ListLoadCasesByComponent returns all load cases (manual+derived) for a component.
func (s *Store) ListLoadCasesByComponent(ctx context.Context, componentID string) ([]model.LoadCase, error) {
	return s.listLoadCases(ctx, `SELECT id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at
		FROM load_cases WHERE component_id=? ORDER BY origin, kind, created_at`, componentID)
}

// ListDerivedCasesByComponent returns only derived cases for a component.
func (s *Store) ListDerivedCasesByComponent(ctx context.Context, componentID string) ([]model.LoadCase, error) {
	return s.listLoadCases(ctx, `SELECT id,project_id,component_id,kind,load_type,magnitude,direction,origin,source_event_id,note,created_at
		FROM load_cases WHERE component_id=? AND origin='manual' ORDER BY kind, created_at`, componentID)
}

func (s *Store) listLoadCases(ctx context.Context, q string, args ...any) ([]model.LoadCase, error) {
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list load_cases: %w", err)
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
