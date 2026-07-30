package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kejixiaoliang/ClassPilot/internal/database"
	"github.com/kejixiaoliang/ClassPilot/internal/workspace"
)

type BuildInfo struct {
	Version string
	Commit  string
}

type App struct {
	Build BuildInfo
}

func New(build BuildInfo) *App {
	return &App{Build: build}
}

type Runtime struct {
	Build    BuildInfo
	Paths    workspace.Paths
	DB       *sql.DB
	ReadOnly bool
	release  func() error
}

func Prepare(ctx context.Context, root string, build BuildInfo) (*Runtime, error) {
	paths, err := workspace.Ensure(root)
	if err != nil {
		return nil, err
	}
	release, err := workspace.AcquireLock(paths.Root)
	if err != nil {
		return nil, err
	}

	db, err := database.Open(paths.Database)
	if err != nil {
		release()
		return nil, err
	}
	if err = database.Migrate(ctx, db); err != nil {
		db.Close()
		release()
		return nil, err
	}

	runtime := &Runtime{
		Build:   build,
		Paths:   paths,
		DB:      db,
		release: release,
	}
	if err = database.IntegrityCheck(ctx, db); err != nil {
		runtime.ReadOnly = true
	}
	return runtime, nil
}

func (runtime *Runtime) Close() error {
	var dbErr, releaseErr error
	if runtime.DB != nil {
		dbErr = runtime.DB.Close()
	}
	if runtime.release != nil {
		releaseErr = runtime.release()
	}
	if dbErr != nil {
		return fmt.Errorf("关闭数据库：%w", dbErr)
	}
	if releaseErr != nil {
		return fmt.Errorf("释放工作台锁：%w", releaseErr)
	}
	return nil
}
