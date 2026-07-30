package customfields

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (repository *Repository) Insert(ctx context.Context, item FieldDefinition) error {
	options, _ := json.Marshal(item.Options)
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO custom_fields (
			id, class_id, name, field_type, description, required, enabled,
			show_in_list, options_json, sort_order, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.ClassID, item.Name, item.Type, item.Description,
		boolInt(item.Required), boolInt(item.Enabled), boolInt(item.ShowInList),
		string(options), item.SortOrder, formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	if err != nil {
		return fmt.Errorf("新增自定义字段：%w", err)
	}
	return nil
}

func (repository *Repository) Get(ctx context.Context, id string) (*FieldDefinition, error) {
	item, err := scanField(repository.db.QueryRowContext(ctx, `
		SELECT `+fieldColumns+` FROM custom_fields WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询自定义字段：%w", err)
	}
	return &item, nil
}

func (repository *Repository) List(ctx context.Context, classID string, includeDisabled bool) ([]FieldDefinition, error) {
	query := `SELECT ` + fieldColumns + ` FROM custom_fields WHERE class_id = ?`
	if !includeDisabled {
		query += ` AND enabled = 1`
	}
	query += ` ORDER BY sort_order ASC, created_at ASC`
	rows, err := repository.db.QueryContext(ctx, query, classID)
	if err != nil {
		return nil, fmt.Errorf("查询自定义字段列表：%w", err)
	}
	defer rows.Close()

	items := make([]FieldDefinition, 0)
	for rows.Next() {
		item, scanErr := scanField(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repository *Repository) Update(ctx context.Context, item FieldDefinition) error {
	options, _ := json.Marshal(item.Options)
	result, err := repository.db.ExecContext(ctx, `
		UPDATE custom_fields
		SET name = ?, description = ?, required = ?, enabled = ?,
			show_in_list = ?, options_json = ?, sort_order = ?, updated_at = ?
		WHERE id = ?
	`, item.Name, item.Description, boolInt(item.Required), boolInt(item.Enabled),
		boolInt(item.ShowInList), string(options), item.SortOrder,
		formatTime(item.UpdatedAt), item.ID)
	if err != nil {
		return fmt.Errorf("更新自定义字段：%w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (repository *Repository) NextSortOrder(ctx context.Context, classID string) (int, error) {
	var value int
	err := repository.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(sort_order), -1) + 1 FROM custom_fields WHERE class_id = ?
	`, classID).Scan(&value)
	return value, err
}

const fieldColumns = `
	id, class_id, name, field_type, description, required, enabled,
	show_in_list, options_json, sort_order, created_at, updated_at
`

type scanner interface {
	Scan(dest ...any) error
}

func scanField(row scanner) (FieldDefinition, error) {
	var item FieldDefinition
	var required, enabled, showInList int
	var optionsJSON, createdAt, updatedAt string
	err := row.Scan(
		&item.ID, &item.ClassID, &item.Name, &item.Type, &item.Description,
		&required, &enabled, &showInList, &optionsJSON, &item.SortOrder,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return FieldDefinition{}, err
	}
	item.Required = required == 1
	item.Enabled = enabled == 1
	item.ShowInList = showInList == 1
	if err = json.Unmarshal([]byte(optionsJSON), &item.Options); err != nil {
		return FieldDefinition{}, err
	}
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return FieldDefinition{}, err
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	return item, err
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
