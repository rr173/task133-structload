package selfcheck

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/service"
	"task133-structload/internal/store"
)

// TestProbeCreateProject calls the service directly (no HTTP) to isolate hangs.
func TestProbeCreateProject(t *testing.T) {
	tmp := t.TempDir() + "/probe.db"
	st, err := store.Open(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := service.New(st, clock.NewFake(time.UnixMilli(1740000000000)))
	p, err := svc.CreateProject(context.Background(), service.CreateProjectInput{
		Code: "X", Name: "X", Exposure: "C", Importance: "II", WindSpeed: 400, GroundSnow: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, "probe: created project", p.ID)
}
