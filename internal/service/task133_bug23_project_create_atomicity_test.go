package service

import (
	"context"
	"path/filepath"
	"testing"

	"task133-structload/internal/clock"
	"task133-structload/internal/store"
)

func TestBug23_ProjectCreationRollsBackWhenAuditWriteFails(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	ctx := context.Background()
	if _, err := st.DB().ExecContext(ctx, `CREATE TRIGGER reject_project_create_event
		BEFORE INSERT ON event_log WHEN NEW.event_type='create_project'
		BEGIN SELECT RAISE(ABORT, 'audit sink unavailable'); END`); err != nil {
		t.Fatal(err)
	}
	svc := New(st, clock.Real{})
	if _, err := svc.CreateProject(ctx, CreateProjectInput{
		Code: "P-23", Name: "atomic project", Exposure: "C", Importance: "II", WindSpeed: 380,
	}); err == nil {
		t.Fatal("project creation should fail when its audit event cannot be written")
	}
	projects, err := svc.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Fatalf("failed creation left persisted projects: %+v", projects)
	}
}
