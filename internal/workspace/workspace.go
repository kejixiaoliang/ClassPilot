package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var ErrAlreadyRunning = errors.New("工作台目录已被另一个 ClassPilot 实例使用")

type Paths struct {
	Root             string
	Data             string
	Attachments      string
	Imports          string
	Exports          string
	Backups          string
	Logs             string
	Database         string
	DatabaseRelative string
}

func Ensure(root string) (Paths, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Paths{}, fmt.Errorf("解析工作台目录：%w", err)
	}

	paths := Paths{
		Root:             absoluteRoot,
		Data:             filepath.Join(absoluteRoot, "data"),
		Attachments:      filepath.Join(absoluteRoot, "attachments"),
		Imports:          filepath.Join(absoluteRoot, "imports"),
		Exports:          filepath.Join(absoluteRoot, "exports"),
		Backups:          filepath.Join(absoluteRoot, "backups"),
		Logs:             filepath.Join(absoluteRoot, "logs"),
		DatabaseRelative: filepath.Join("data", "classpilot.sqlite"),
	}
	paths.Database = filepath.Join(paths.Root, paths.DatabaseRelative)

	for _, dir := range []string{
		paths.Data,
		paths.Attachments,
		paths.Imports,
		paths.Exports,
		paths.Backups,
		paths.Logs,
	} {
		if mkdirErr := os.MkdirAll(dir, 0o755); mkdirErr != nil {
			return Paths{}, fmt.Errorf("创建工作台目录 %s：%w", dir, mkdirErr)
		}
	}

	return paths, nil
}

func AcquireLock(root string) (func() error, error) {
	lockPath := filepath.Join(root, ".classpilot.lock")
	file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil, ErrAlreadyRunning
	}
	if err != nil {
		return nil, fmt.Errorf("创建工作台锁：%w", err)
	}

	if _, err = fmt.Fprintf(file, "pid=%d\nstarted_at=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339)); err != nil {
		file.Close()
		os.Remove(lockPath)
		return nil, fmt.Errorf("写入工作台锁：%w", err)
	}

	released := false
	release := func() error {
		if released {
			return nil
		}
		released = true
		closeErr := file.Close()
		removeErr := os.Remove(lockPath)
		if closeErr != nil {
			return closeErr
		}
		if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return removeErr
		}
		return nil
	}

	return release, nil
}
