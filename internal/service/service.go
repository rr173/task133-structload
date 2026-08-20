package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/idlib"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

// Service is the business-logic layer over the Store. It owns validation,
// derivation (wind/snow/live reduction), check computation, overrides, and
// the ReconcileAll recovery path. All authoritative writes append an event.
type Service struct {
	store *store.Store
	clk   clock.Clock
}

// New constructs a Service backed by the given store and clock.
func New(s *store.Store, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.Real{}
	}
	return &Service{store: s, clk: clk}
}

// now returns the service clock's current time.
func (svc *Service) now() time.Time { return svc.clk.Now() }

// appendEvent is a helper that records an event with the next project seq.
// Runs in its own transaction; all writes inside use the tx (NOT s.db) to
// avoid SetMaxOpenConns(1) deadlock.
func (svc *Service) appendEvent(ctx context.Context, projectID, eventType string, payload []byte) error {
	return svc.store.InTx(ctx, func(tx store.DBTX) error {
		seq, err := svc.store.NextEventSeq(ctx, tx, projectID)
		if err != nil {
			return err
		}
		e := model.EventLog{
			ID:        idlib.NewID("evt"),
			ProjectID: projectID,
			Seq:       seq,
			EventType: eventType,
			Payload:   "",
			CreatedAt: svc.now(),
		}
		return svc.store.AppendEventInTx(ctx, tx, e)
	})
}

// wrapErr attaches a message to a sentinel error while preserving identity.
func wrapErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// isNotFound reports whether err is (or wraps) model.ErrNotFound.
func isNotFound(err error) bool { return errors.Is(err, model.ErrNotFound) }
