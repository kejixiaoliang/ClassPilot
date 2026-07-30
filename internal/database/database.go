package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("解析数据库路径：%w", err)
	}
	if err = os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, fmt.Errorf("创建数据库目录：%w", err)
	}

	urlPath := filepath.ToSlash(absolutePath)
	if filepath.VolumeName(absolutePath) != "" && !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	dsn := (&url.URL{
		Scheme:   "file",
		Path:     urlPath,
		RawQuery: "_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)",
	}).String()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库：%w", err)
	}
	db.SetMaxOpenConns(1)

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接数据库：%w", err)
	}
	return db, nil
}

func IntegrityCheck(ctx context.Context, db *sql.DB) error {
	var result string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&result); err != nil {
		return fmt.Errorf("执行数据库完整性检查：%w", err)
	}
	if result != "ok" {
		return fmt.Errorf("数据库完整性检查未通过：%s", result)
	}
	return nil
}

func Backup(ctx context.Context, db *sql.DB, destination string) error {
	absoluteDestination, err := filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("解析备份路径：%w", err)
	}
	if err = os.MkdirAll(filepath.Dir(absoluteDestination), 0o755); err != nil {
		return fmt.Errorf("创建备份目录：%w", err)
	}
	if _, err = os.Stat(absoluteDestination); err == nil {
		return fmt.Errorf("备份文件已存在：%s", filepath.Base(absoluteDestination))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查备份文件：%w", err)
	}

	escaped := strings.ReplaceAll(filepath.ToSlash(absoluteDestination), "'", "''")
	if _, err = db.ExecContext(ctx, "VACUUM INTO '"+escaped+"'"); err != nil {
		return fmt.Errorf("创建数据库备份：%w", err)
	}

	backupDB, err := Open(absoluteDestination)
	if err != nil {
		os.Remove(absoluteDestination)
		return err
	}
	defer backupDB.Close()
	if err = IntegrityCheck(ctx, backupDB); err != nil {
		backupDB.Close()
		os.Remove(absoluteDestination)
		return fmt.Errorf("校验数据库备份：%w", err)
	}
	return nil
}
