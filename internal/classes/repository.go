package classes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) Insert(ctx context.Context, class Class) error {
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO classes (
			id, name, stage, grade, school_year, status, notes,
			created_at, updated_at, archived_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, class.ID, class.Name, class.Stage, class.Grade, class.SchoolYear,
		class.Status, class.Notes, formatTime(class.CreatedAt),
		formatTime(class.UpdatedAt), formatOptionalTime(class.ArchivedAt))
	if err != nil {
		return fmt.Errorf("新增班级：%w", err)
	}
	return nil
}

func (repository *Repository) Current(ctx context.Context) (*Class, error) {
	return repository.queryOne(ctx, `SELECT `+classColumns+` FROM classes WHERE status = 'current' LIMIT 1`)
}

func (repository *Repository) Get(ctx context.Context, id string) (*Class, error) {
	return repository.queryOne(ctx, `SELECT `+classColumns+` FROM classes WHERE id = ?`, id)
}

func (repository *Repository) ListArchived(ctx context.Context) ([]Class, error) {
	rows, err := repository.db.QueryContext(ctx, `
		SELECT `+classColumns+`
		FROM classes
		WHERE status = 'archived'
		ORDER BY archived_at DESC, updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询历史班级：%w", err)
	}
	defer rows.Close()

	var items []Class
	for rows.Next() {
		item, scanErr := scanClass(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) Update(ctx context.Context, class Class) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE classes
		SET name = ?, stage = ?, grade = ?, school_year = ?, notes = ?, updated_at = ?
		WHERE id = ? AND status = 'current'
	`, class.Name, class.Stage, class.Grade, class.SchoolYear, class.Notes,
		formatTime(class.UpdatedAt), class.ID)
	if err != nil {
		return fmt.Errorf("更新班级：%w", err)
	}
	return requireAffected(result)
}

func (repository *Repository) Archive(ctx context.Context, id string, at time.Time) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE classes
		SET status = 'archived', archived_at = ?, updated_at = ?
		WHERE id = ? AND status = 'current'
	`, formatTime(at), formatTime(at), id)
	if err != nil {
		return fmt.Errorf("归档班级：%w", err)
	}
	return requireAffected(result)
}

func (repository *Repository) Restore(ctx context.Context, id string, at time.Time) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE classes
		SET status = 'current', archived_at = NULL, updated_at = ?
		WHERE id = ? AND status = 'archived'
	`, formatTime(at), id)
	if err != nil {
		return fmt.Errorf("恢复班级：%w", err)
	}
	return requireAffected(result)
}

const classColumns = `
	id, name, stage, grade, school_year, status, notes,
	created_at, updated_at, archived_at
`

type scanner interface {
	Scan(dest ...any) error
}

func scanClass(row scanner) (Class, error) {
	var item Class
	var createdAt, updatedAt string
	var archivedAt sql.NullString
	err := row.Scan(
		&item.ID, &item.Name, &item.Stage, &item.Grade, &item.SchoolYear,
		&item.Status, &item.Notes, &createdAt, &updatedAt, &archivedAt,
	)
	if err != nil {
		return Class{}, err
	}
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Class{}, err
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Class{}, err
	}
	if archivedAt.Valid {
		value, parseErr := time.Parse(time.RFC3339Nano, archivedAt.String)
		if parseErr != nil {
			return Class{}, parseErr
		}
		item.ArchivedAt = &value
	}
	return item, nil
}

func (repository *Repository) queryOne(ctx context.Context, query string, args ...any) (*Class, error) {
	item, err := scanClass(repository.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询班级：%w", err)
	}
	return &item, nil
}

func requireAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}
