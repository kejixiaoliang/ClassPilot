package database_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/database"
)

func TestMigrateCreatesInitialSchema(t *testing.T) {
	db := openTestDB(t)

	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	for _, table := range []string{
		"schema_migrations",
		"classes",
		"students",
		"custom_fields",
		"student_field_values",
		"attachments",
		"audit_logs",
	} {
		if !tableExists(t, db, table) {
			t.Fatalf("数据表未创建：%s", table)
		}
	}
}

func TestOnlyOneCurrentClassIsAllowed(t *testing.T) {
	db := openTestDB(t)
	mustMigrate(t, db)

	_, err := db.Exec(`
		INSERT INTO classes (id, name, stage, grade, school_year, status, created_at, updated_at)
		VALUES ('class-1', '一班', '初中', '七年级', '2026-2027', 'current', '2026-07-31T00:00:00Z', '2026-07-31T00:00:00Z')
	`)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = db.Exec(`
		INSERT INTO classes (id, name, stage, grade, school_year, status, created_at, updated_at)
		VALUES ('class-2', '二班', '初中', '七年级', '2026-2027', 'current', '2026-07-31T00:00:00Z', '2026-07-31T00:00:00Z')
	`); err == nil {
		t.Fatal("数据库应拒绝第二个当前班级")
	}
}

func TestIntegrityCheckReportsHealthyDatabase(t *testing.T) {
	db := openTestDB(t)
	mustMigrate(t, db)

	if err := database.IntegrityCheck(context.Background(), db); err != nil {
		t.Fatalf("完整性检查失败：%v", err)
	}
}

func TestBackupCreatesReadableSnapshot(t *testing.T) {
	db := openTestDB(t)
	mustMigrate(t, db)
	_, err := db.Exec(`
		INSERT INTO classes (id, name, stage, grade, school_year, status, created_at, updated_at)
		VALUES ('class-1', '七年级一班', '初中', '七年级', '2026-2027', 'current', '2026-07-31T00:00:00Z', '2026-07-31T00:00:00Z')
	`)
	if err != nil {
		t.Fatal(err)
	}

	backupPath := filepath.Join(t.TempDir(), "snapshot.sqlite")
	if err = database.Backup(context.Background(), db, backupPath); err != nil {
		t.Fatal(err)
	}

	backupDB, err := database.Open(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	defer backupDB.Close()

	var name string
	if err = backupDB.QueryRow(`SELECT name FROM classes WHERE id = 'class-1'`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "七年级一班" {
		t.Fatalf("备份内容=%q", name)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustMigrate(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
		name,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count == 1
}
