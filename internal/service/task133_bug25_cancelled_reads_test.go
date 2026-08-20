package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"task133-structload/internal/clock"
	"task133-structload/internal/store"
)

func TestBug25_CancelledEntityReadsPropagateContextErrors(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "cancel.db"))
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := New(st, clock.Real{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := svc.GetProject(ctx, "missing-project"); !errors.Is(err, context.Canceled) { t.Fatalf("GetProject error = %v, want context.Canceled", err) }
	if _, err := svc.GetComponent(ctx, "missing-component"); !errors.Is(err, context.Canceled) { t.Fatalf("GetComponent error = %v, want context.Canceled", err) }
}
