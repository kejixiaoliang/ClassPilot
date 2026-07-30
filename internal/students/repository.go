package students

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) Insert(ctx context.Context, item Student) error {
	_, err := repository.db.ExecContext(ctx, `
		INSERT INTO students (
			id, class_id, name, gender, student_number, birth_date,
			enrollment_date, status, notes, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
	`, item.ID, item.ClassID, item.Name, item.Gender, nullableString(item.StudentNumber),
		formatDate(item.BirthDate), formatDate(item.EnrollmentDate), item.Status, item.Notes,
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt))
	if err != nil {
		return fmt.Errorf("新增学生：%w", err)
	}
	return nil
}

func (repository *Repository) Get(ctx context.Context, id string) (*Student, error) {
	item, err := scanStudent(repository.db.QueryRowContext(ctx, `
		SELECT `+studentColumns+` FROM students WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询学生：%w", err)
	}
	return &item, nil
}

func (repository *Repository) Update(ctx context.Context, item Student) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE students
		SET name = ?, gender = ?, student_number = ?, birth_date = ?,
			enrollment_date = ?, status = ?, notes = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Name, item.Gender, nullableString(item.StudentNumber), formatDate(item.BirthDate),
		formatDate(item.EnrollmentDate), item.Status, item.Notes, formatTime(item.UpdatedAt), item.ID)
	if err != nil {
		return fmt.Errorf("更新学生：%w", err)
	}
	return requireAffected(result)
}

func (repository *Repository) SetDeletedAt(ctx context.Context, id string, deletedAt *time.Time) error {
	result, err := repository.db.ExecContext(ctx, `
		UPDATE students SET deleted_at = ?, updated_at = ? WHERE id = ?
	`, formatDateTime(deletedAt), formatTime(time.Now().UTC()), id)
	if err != nil {
		return fmt.Errorf("更新学生回收站状态：%w", err)
	}
	return requireAffected(result)
}

func (repository *Repository) List(ctx context.Context, query Query) (Page, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 200 {
		query.PageSize = 20
	}

	conditions := []string{"class_id = ?"}
	args := []any{query.ClassID}
	if !query.IncludeTrashed {
		conditions = append(conditions, "deleted_at IS NULL")
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		conditions = append(conditions, "(name LIKE ? ESCAPE '\\' OR student_number LIKE ? ESCAPE '\\')")
		search = "%" + escapeLike(search) + "%"
		args = append(args, search, search)
	}
	where := strings.Join(conditions, " AND ")

	var total int
	if err := repository.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM students WHERE `+where, args...).Scan(&total); err != nil {
		return Page{}, fmt.Errorf("统计学生：%w", err)
	}

	sortColumn := map[string]string{
		"name": "name", "studentNumber": "student_number", "createdAt": "created_at",
	}[query.SortBy]
	if sortColumn == "" {
		sortColumn = "created_at"
	}
	sortOrder := "DESC"
	if strings.EqualFold(query.SortOrder, "asc") {
		sortOrder = "ASC"
	}

	listArgs := append(append([]any{}, args...), query.PageSize, (query.Page-1)*query.PageSize)
	rows, err := repository.db.QueryContext(ctx, `
		SELECT `+studentColumns+`
		FROM students
		WHERE `+where+`
		ORDER BY `+sortColumn+` `+sortOrder+`, id ASC
		LIMIT ? OFFSET ?
	`, listArgs...)
	if err != nil {
		return Page{}, fmt.Errorf("查询学生列表：%w", err)
	}
	defer rows.Close()

	items := make([]Student, 0)
	for rows.Next() {
		item, scanErr := scanStudent(rows)
		if scanErr != nil {
			return Page{}, scanErr
		}
		items = append(items, item)
	}
	return Page{Items: items, Total: total, Page: query.Page, PageSize: query.PageSize}, rows.Err()
}

const studentColumns = `
	id, class_id, name, gender, COALESCE(student_number, ''), birth_date,
	enrollment_date, status, notes, created_at, updated_at, deleted_at
`

type scanner interface {
	Scan(dest ...any) error
}

func scanStudent(row scanner) (Student, error) {
	var item Student
	var birthDate, enrollmentDate, deletedAt sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(
		&item.ID, &item.ClassID, &item.Name, &item.Gender, &item.StudentNumber,
		&birthDate, &enrollmentDate, &item.Status, &item.Notes,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return Student{}, err
	}
	item.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Student{}, err
	}
	item.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Student{}, err
	}
	item.BirthDate, err = parseOptionalDate(birthDate)
	if err != nil {
		return Student{}, err
	}
	item.EnrollmentDate, err = parseOptionalDate(enrollmentDate)
	if err != nil {
		return Student{}, err
	}
	item.DeletedAt, err = parseOptionalTime(deletedAt)
	return item, err
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

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func formatDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}

func formatDateTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseOptionalDate(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value.String, time.Local)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}
