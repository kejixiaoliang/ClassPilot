package audit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSensitiveContent = errors.New("操作记录摘要包含敏感内容")
	phonePattern        = regexp.MustCompile(`(?:^|\D)1[3-9]\d{9}(?:\D|$)`)
	idCardPattern       = regexp.MustCompile(`(?:^|\D)\d{17}[\dXx](?:\D|$)`)
	windowsPathPattern  = regexp.MustCompile(`(?i)[a-z]:\\`)
)

type Entry struct {
	ID         string
	Action     string
	EntityType string
	EntityID   string
	Summary    string
	CreatedAt  time.Time
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (service *Service) Record(ctx context.Context, entry Entry) error {
	if containsSensitiveContent(entry.Summary) {
		return ErrSensitiveContent
	}
	if service.db == nil {
		return errors.New("操作记录数据库未配置")
	}
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	if strings.TrimSpace(entry.Action) == "" ||
		strings.TrimSpace(entry.EntityType) == "" ||
		strings.TrimSpace(entry.EntityID) == "" {
		return errors.New("操作类型、对象类型和对象标识不能为空")
	}

	_, err := service.db.ExecContext(ctx, `
		INSERT INTO audit_logs (id, action, entity_type, entity_id, summary, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, entry.ID, entry.Action, entry.EntityType, entry.EntityID, entry.Summary, entry.CreatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("写入操作记录：%w", err)
	}
	return nil
}

func containsSensitiveContent(value string) bool {
	return phonePattern.MatchString(value) ||
		idCardPattern.MatchString(value) ||
		windowsPathPattern.MatchString(value)
}
