package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task133-structload/internal/clock"
	"task133-structload/internal/model"
	"task133-structload/internal/store"
)

func TestBug22_ProjectListingPreservesCreationOrder(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "projects.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := New(st, clock.NewFake(time.UnixMilli(1740000000000)))
	ctx := context.Background()
	want := make([]string, 0, 3)
	for _, code := range []string{"A-first", "B-second", "C-third"} {
		p, err := svc.CreateProject(ctx, CreateProjectInput{
			Code: code, Name: code, Exposure: "C", Importance: "II", WindSpeed: 380,
		})
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, p.ID)
	}
	projects, err := svc.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != len(want) {
		t.Fatalf("project count = %d, want %d", len(projects), len(want))
	}
	for i, p := range projects {
		if p.ID != want[i] {
			t.Fatalf("project order = %v, want creation order %v", projectIDs(projects), want)
		}
	}
}

func projectIDs(projects []model.Project) []string {
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	return ids
}
