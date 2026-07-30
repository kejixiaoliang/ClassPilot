package students_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/classes"
	"github.com/kejixiaoliang/ClassPilot/internal/database"
	"github.com/kejixiaoliang/ClassPilot/internal/students"
)

func TestStudentNumberUniqueWithinClass(t *testing.T) {
	service, classID := newTestService(t)
	input := students.CreateInput{Name: "张三", StudentNumber: "2026001"}

	if _, err := service.Create(context.Background(), classID, input); err != nil {
		t.Fatal(err)
	}
	input.Name = "李四"
	if _, err := service.Create(context.Background(), classID, input); !errors.Is(err, students.ErrStudentNumberExists) {
		t.Fatalf("期望 ErrStudentNumberExists，得到 %v", err)
	}
}

func TestAgeIsCalculatedFromBirthDate(t *testing.T) {
	birthDate := time.Date(2012, time.August, 10, 0, 0, 0, 0, time.Local)
	beforeBirthday := time.Date(2026, time.August, 9, 0, 0, 0, 0, time.Local)
	onBirthday := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.Local)

	if got := students.AgeAt(birthDate, beforeBirthday); got != 13 {
		t.Fatalf("生日前年龄=%d，期望=13", got)
	}
	if got := students.AgeAt(birthDate, onBirthday); got != 14 {
		t.Fatalf("生日当天年龄=%d，期望=14", got)
	}
}

func TestListExcludesTrashedByDefault(t *testing.T) {
	service, classID := newTestService(t)
	created, err := service.Create(context.Background(), classID, students.CreateInput{Name: "张三"})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Trash(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}

	page, err := service.List(context.Background(), students.Query{ClassID: classID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("默认列表包含回收站学生：%d", page.Total)
	}
}

func TestTrashThenRestoreStudent(t *testing.T) {
	service, classID := newTestService(t)
	created, err := service.Create(context.Background(), classID, students.CreateInput{Name: "张三"})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Trash(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.Restore(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}

	page, err := service.List(context.Background(), students.Query{ClassID: classID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Items[0].Name != "张三" {
		t.Fatalf("恢复结果=%+v", page)
	}
}

func TestArchivedClassRejectsStudentWrites(t *testing.T) {
	service, classID, classService := newTestServices(t)
	if err := classService.ArchiveCurrent(context.Background(), classID); err != nil {
		t.Fatal(err)
	}

	_, err := service.Create(context.Background(), classID, students.CreateInput{Name: "张三"})
	if !errors.Is(err, students.ErrClassReadOnly) {
		t.Fatalf("期望 ErrClassReadOnly，得到 %v", err)
	}
}

func newTestService(t *testing.T) (*students.Service, string) {
	t.Helper()
	service, classID, _ := newTestServices(t)
	return service, classID
}

func newTestServices(t *testing.T) (*students.Service, string, *classes.Service) {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "students.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	auditService := audit.New(db)
	classService := classes.NewService(classes.NewRepository(db), auditService)
	current, err := classService.CreateCurrent(context.Background(), classes.CreateInput{
		Name: "七年级一班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027",
	})
	if err != nil {
		t.Fatal(err)
	}
	return students.NewService(students.NewRepository(db), classes.NewRepository(db), auditService), current.ID, classService
}
