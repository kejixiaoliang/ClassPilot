package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	migration, err := migrationFiles.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("读取初始数据库迁移：%w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始数据库迁移：%w", err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, string(migration)); err != nil {
		return fmt.Errorf("执行初始数据库迁移：%w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交数据库迁移：%w", err)
	}
	return nil
}
