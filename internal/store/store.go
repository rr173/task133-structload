package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"task133-structload/internal/model"
)

// DBTX is the read interface satisfied by both *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Store is the persistence layer over a single SQLite file.
// SetMaxOpenConns(1): the engine is single-connection; tx-internal reads use tx.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite file and runs migrations. WAL + NORMAL sync.
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	s := &Store{db: db}
	if err := s.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// DB exposes the underlying handle for tests/service helpers.
func (s *Store) DB() *sql.DB { return s.db }

// Migrate applies the schema.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, Schema)
	return err
}

// InTx runs fn inside a single transaction. Reads inside fn must use tx (DBTX),
// never the shared *sql.DB, to avoid SetMaxOpenConns(1) deadlock.
func (s *Store) InTx(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// nowMS returns unix milliseconds for storage columns.
func nowMS(t time.Time) int64 { return t.UnixMilli() }

// fromMS converts stored milliseconds back to a Time.
func fromMS(ms int64) time.Time { return time.UnixMilli(ms).UTC() }

// LoadAll reads all authoritative state into memory maps. Used at startup and
// before ReconcileAll so the engine can rebuild derived state deterministically.
type Snapshot struct {
	Projects     map[string]model.Project
	Levels       map[string]model.Level
	Components   map[string]model.Component
	ManualCases  []model.LoadCase   // origin=manual
	DerivedCases []model.LoadCase   // origin=derived (snapshot before reconcile)
	Combos       map[string]model.LoadCombination
	Overrides    []model.Override
	Events       []model.EventLog
}

func (s *Store) LoadAll(ctx context.Context) (*Snapshot, error) {
	snap := &Snapshot{
		Projects:   map[string]model.Project{},
		Levels:     map[string]model.Level{},
		Components: map[string]model.Component{},
		Combos:     map[string]model.LoadCombination{},
	}
	if err := s.loadProjects(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadLevels(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadComponents(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadLoadCases(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadCombos(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadOverrides(ctx, snap); err != nil {
		return nil, err
	}
	if err := s.loadEvents(ctx, snap); err != nil {
		return nil, err
	}
	return snap, nil
}

// DeleteDerivedAndChecks wipes computed load cases and checks before ReconcileAll.
func (s *Store) DeleteDerivedAndChecks(ctx context.Context) error {
	return s.InTx(ctx, func(tx DBTX) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM checks`); err != nil {
			return fmt.Errorf("delete checks: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM load_cases WHERE origin='derived'`); err != nil {
			return fmt.Errorf("delete derived load_cases: %w", err)
		}
		return nil
	})
}
