package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/workspace"
)

func TestEnsureCreatesPortableWorkspace(t *testing.T) {
	root := t.TempDir()

	paths, err := workspace.Ensure(root)
	if err != nil {
		t.Fatal(err)
	}

	for _, dir := range []string{
		paths.Data,
		paths.Attachments,
		paths.Imports,
		paths.Exports,
		paths.Backups,
		paths.Logs,
	} {
		info, statErr := os.Stat(dir)
		if statErr != nil || !info.IsDir() {
			t.Fatalf("目录未创建：%s", dir)
		}
	}

	wantRelative := filepath.Join("data", "classpilot.sqlite")
	if paths.DatabaseRelative != wantRelative {
		t.Fatalf("数据库相对路径=%q，期望=%q", paths.DatabaseRelative, wantRelative)
	}
	if paths.Database != filepath.Join(paths.Root, wantRelative) {
		t.Fatalf("数据库绝对路径不在工作台内：%s", paths.Database)
	}
}

func TestAcquireLockRejectsSecondWriter(t *testing.T) {
	root := t.TempDir()

	release, err := workspace.AcquireLock(root)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	if _, secondErr := workspace.AcquireLock(root); secondErr == nil {
		t.Fatal("第二个写入实例应被拒绝")
	}
}
