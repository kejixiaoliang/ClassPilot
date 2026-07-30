package audit_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/database"
)

func TestRecordStoresSafeOperationSummary(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "audit.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	service := audit.New(db)
	err = service.Record(context.Background(), audit.Entry{
		Action:     "student.created",
		EntityType: "student",
		EntityID:   "student-1",
		Summary:    "新增学生记录",
	})
	if err != nil {
		t.Fatal(err)
	}

	var count int
	if err = db.QueryRow(`SELECT COUNT(*) FROM audit_logs WHERE entity_id = 'student-1'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("操作记录数量=%d，期望=1", count)
	}
}

func TestRecordRejectsSensitiveSummary(t *testing.T) {
	service := audit.New(nil)
	for _, summary := range []string{
		"联系电话 13812345678",
		"身份证 110101200001011234",
		`附件 C:\Users\teacher\student.pdf`,
	} {
		err := service.Record(context.Background(), audit.Entry{
			Action:     "student.updated",
			EntityType: "student",
			EntityID:   "student-1",
			Summary:    summary,
		})
		if !errors.Is(err, audit.ErrSensitiveContent) {
			t.Fatalf("摘要 %q 应被拒绝，得到 %v", summary, err)
		}
	}
}
