package classes_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/classes"
	"github.com/kejixiaoliang/ClassPilot/internal/database"
)

func TestOnlyOneCurrentClass(t *testing.T) {
	service := newTestService(t)

	_, err := service.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "七年级一班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "七年级二班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027",
	})
	if !errors.Is(err, classes.ErrCurrentExists) {
		t.Fatalf("期望 ErrCurrentExists，得到 %v", err)
	}
}

func TestArchivedClassIsReadOnly(t *testing.T) {
	service := newTestService(t)
	created, err := service.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "七年级一班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ArchiveCurrent(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}

	_, err = service.Update(context.Background(), created.ID, classes.UpdateInput{Name: "修改后的班级"})
	if !errors.Is(err, classes.ErrArchivedReadOnly) {
		t.Fatalf("期望 ErrArchivedReadOnly，得到 %v", err)
	}
}

func TestRestoreRequiresNoCurrentClass(t *testing.T) {
	service := newTestService(t)
	archived, err := service.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "往届一班", Stage: "初中", Grade: "九年级", SchoolYear: "2025-2026",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ArchiveCurrent(context.Background(), archived.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "当前一班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027",
	}); err != nil {
		t.Fatal(err)
	}

	err = service.RestoreAsCurrent(context.Background(), archived.ID)
	if !errors.Is(err, classes.ErrCurrentExists) {
		t.Fatalf("期望 ErrCurrentExists，得到 %v", err)
	}
}

func newTestService(t *testing.T) *classes.Service {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "classes.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return classes.NewService(classes.NewRepository(db), audit.New(db))
}
