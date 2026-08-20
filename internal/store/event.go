package store

import (
	"context"
	"fmt"

	"task133-structload/internal/model"
)

// AppendEvent inserts an event log row with a project-scoped sequence number.
// Uses the shared *sql.DB: only safe OUTSIDE a transaction (see AppendEventInTx for tx use).
func (s *Store) AppendEvent(ctx context.Context, e model.EventLog) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO event_log (id,project_id,seq,event_type,payload,created_at)
		VALUES(?,?,?,?,?,?)`,
		e.ID, e.ProjectID, e.Seq, e.EventType, e.Payload, nowMS(e.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert event_log: %w", err)
	}
	return nil
}

// AppendEventInTx inserts an event within an existing transaction.
// MUST be used when already inside InTx to avoid SetMaxOpenConns(1) deadlock.
func (s *Store) AppendEventInTx(ctx context.Context, tx DBTX, e model.EventLog) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO event_log (id,project_id,seq,event_type,payload,created_at)
		VALUES(?,?,?,?,?,?)`,
		e.ID, e.ProjectID, e.Seq, e.EventType, e.Payload, nowMS(e.CreatedAt))
	if err != nil {
		return fmt.Errorf("insert event_log: %w", err)
	}
	return nil
}

// NextEventSeq returns the next project-scoped sequence number in a transaction.
func (s *Store) NextEventSeq(ctx context.Context, tx DBTX, projectID string) (int64, error) {
	var seq int64
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(seq),0)+1 FROM event_log`).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("next event seq: %w", err)
	}
	return seq, nil
}

// ListEvents returns the event log for a project.
func (s *Store) ListEvents(ctx context.Context, projectID string) ([]model.EventLog, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,project_id,seq,event_type,payload,created_at
		FROM event_log WHERE project_id=? ORDER BY seq`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var out []model.EventLog
	for rows.Next() {
		var e model.EventLog
		var ms int64
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Seq, &e.EventType, &e.Payload, &ms); err != nil {
			return nil, err
		}
		e.CreatedAt = fromMS(ms)
		out = append(out, e)
	}
	return out, rows.Err()
}
