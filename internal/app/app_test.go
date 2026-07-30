package app_test

import (
	"context"
	"os"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/app"
)

func TestNewPreservesBuildInfo(t *testing.T) {
	got := app.New(app.BuildInfo{Version: "0.1.0", Commit: "abc123"})

	if got.Build.Version != "0.1.0" || got.Build.Commit != "abc123" {
		t.Fatalf("构建信息未保留：%+v", got.Build)
	}
}

func TestPrepareCreatesMigratedDatabase(t *testing.T) {
	root := t.TempDir()

	runtime, err := app.Prepare(context.Background(), root, app.BuildInfo{Version: "0.1.0"})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()

	if _, err = os.Stat(runtime.Paths.Database); err != nil {
		t.Fatalf("数据库未创建：%v", err)
	}

	var migrations int
	if err = runtime.DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrations); err != nil {
		t.Fatal(err)
	}
	if migrations != 1 {
		t.Fatalf("迁移记录=%d，期望=1", migrations)
	}
}
