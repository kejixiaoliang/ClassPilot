package customfields_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/classes"
	"github.com/kejixiaoliang/ClassPilot/internal/customfields"
	"github.com/kejixiaoliang/ClassPilot/internal/database"
)

func TestCreateRejectsUnsupportedFieldType(t *testing.T) {
	service, classID := newTestService(t)

	_, err := service.Create(context.Background(), customfields.CreateInput{
		ClassID: classID, Name: "未知字段", Type: "file",
	})
	if !errors.Is(err, customfields.ErrUnsupportedType) {
		t.Fatalf("期望 ErrUnsupportedType，得到 %v", err)
	}
}

func TestRequiredFieldRejectsEmptyValue(t *testing.T) {
	service, classID := newTestService(t)
	field, err := service.Create(context.Background(), customfields.CreateInput{
		ClassID: classID, Name: "宿舍号", Type: customfields.TypeText, Required: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	issues, err := service.ValidateValues(context.Background(), classID, map[string]any{field.ID: ""})
	if err != nil {
		t.Fatal(err)
	}
	if issues[field.ID] != "该字段为必填项" {
		t.Fatalf("校验结果=%v", issues)
	}
}

func TestSingleChoiceRejectsUnknownOption(t *testing.T) {
	service, classID := newTestService(t)
	field, err := service.Create(context.Background(), customfields.CreateInput{
		ClassID: classID, Name: "住宿情况", Type: customfields.TypeSingleChoice,
		Options: []string{"走读", "住校"},
	})
	if err != nil {
		t.Fatal(err)
	}

	issues, err := service.ValidateValues(context.Background(), classID, map[string]any{field.ID: "校外租住"})
	if err != nil {
		t.Fatal(err)
	}
	if issues[field.ID] != "选项不在允许范围内" {
		t.Fatalf("校验结果=%v", issues)
	}
}

func TestDisablePreservesFieldDefinition(t *testing.T) {
	service, classID := newTestService(t)
	field, err := service.Create(context.Background(), customfields.CreateInput{
		ClassID: classID, Name: "兴趣爱好", Type: customfields.TypeText,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Disable(context.Background(), field.ID); err != nil {
		t.Fatal(err)
	}

	items, err := service.List(context.Background(), classID, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Enabled {
		t.Fatalf("停用字段未保留：%+v", items)
	}
}

func newTestService(t *testing.T) (*customfields.Service, string) {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "fields.sqlite"))
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
	return customfields.NewService(customfields.NewRepository(db), classes.NewRepository(db), auditService), current.ID
}
