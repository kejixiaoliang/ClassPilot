package customfields

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/classes"
)

type Service struct {
	repository *Repository
	classes    *classes.Repository
	audit      *audit.Service
	now        func() time.Time
}

func NewService(repository *Repository, classRepository *classes.Repository, auditService *audit.Service) *Service {
	return &Service{
		repository: repository,
		classes:    classRepository,
		audit:      auditService,
		now:        time.Now,
	}
}

func (service *Service) Create(ctx context.Context, input CreateInput) (FieldDefinition, error) {
	if !supportedType(input.Type) {
		return FieldDefinition{}, ErrUnsupportedType
	}
	if strings.TrimSpace(input.ClassID) == "" || strings.TrimSpace(input.Name) == "" {
		return FieldDefinition{}, ErrInvalidInput
	}
	if err := service.ensureWritableClass(ctx, input.ClassID); err != nil {
		return FieldDefinition{}, err
	}
	options := normalizeOptions(input.Options)
	if (input.Type == TypeSingleChoice || input.Type == TypeMultiChoice) && len(options) == 0 {
		return FieldDefinition{}, fmt.Errorf("%w：选项字段至少需要一个选项", ErrInvalidInput)
	}
	sortOrder, err := service.repository.NextSortOrder(ctx, input.ClassID)
	if err != nil {
		return FieldDefinition{}, err
	}
	now := service.now().UTC()
	item := FieldDefinition{
		ID: uuid.NewString(), ClassID: input.ClassID, Name: strings.TrimSpace(input.Name),
		Type: input.Type, Description: strings.TrimSpace(input.Description),
		Required: input.Required, Enabled: true, ShowInList: input.ShowInList,
		Options: options, SortOrder: sortOrder, CreatedAt: now, UpdatedAt: now,
	}
	if err = service.repository.Insert(ctx, item); err != nil {
		return FieldDefinition{}, err
	}
	service.record(ctx, "custom_field.created", item.ID, "创建自定义字段")
	return item, nil
}

func (service *Service) List(ctx context.Context, classID string, includeDisabled bool) ([]FieldDefinition, error) {
	return service.repository.List(ctx, classID, includeDisabled)
}

func (service *Service) Update(ctx context.Context, id string, input UpdateInput) (FieldDefinition, error) {
	item, err := service.repository.Get(ctx, id)
	if err != nil {
		return FieldDefinition{}, err
	}
	if item == nil {
		return FieldDefinition{}, ErrNotFound
	}
	if err = service.ensureWritableClass(ctx, item.ClassID); err != nil {
		return FieldDefinition{}, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return FieldDefinition{}, ErrInvalidInput
	}
	item.Name = strings.TrimSpace(input.Name)
	item.Description = strings.TrimSpace(input.Description)
	item.Required = input.Required
	item.ShowInList = input.ShowInList
	if item.Type == TypeSingleChoice || item.Type == TypeMultiChoice {
		item.Options = normalizeOptions(input.Options)
		if len(item.Options) == 0 {
			return FieldDefinition{}, ErrInvalidInput
		}
	}
	item.UpdatedAt = service.now().UTC()
	if err = service.repository.Update(ctx, *item); err != nil {
		return FieldDefinition{}, err
	}
	service.record(ctx, "custom_field.updated", item.ID, "更新自定义字段")
	return *item, nil
}

func (service *Service) Disable(ctx context.Context, id string) error {
	item, err := service.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrNotFound
	}
	if err = service.ensureWritableClass(ctx, item.ClassID); err != nil {
		return err
	}
	item.Enabled = false
	item.UpdatedAt = service.now().UTC()
	if err = service.repository.Update(ctx, *item); err != nil {
		return err
	}
	service.record(ctx, "custom_field.disabled", item.ID, "停用自定义字段并保留历史值")
	return nil
}

func (service *Service) ValidateValues(ctx context.Context, classID string, values map[string]any) (map[string]string, error) {
	fields, err := service.repository.List(ctx, classID, false)
	if err != nil {
		return nil, err
	}
	issues := make(map[string]string)
	for _, field := range fields {
		value, exists := values[field.ID]
		if !exists || isEmpty(value) {
			if field.Required {
				issues[field.ID] = "该字段为必填项"
			}
			continue
		}
		if issue := validateValue(field, value); issue != "" {
			issues[field.ID] = issue
		}
	}
	return issues, nil
}

func (service *Service) ensureWritableClass(ctx context.Context, classID string) error {
	item, err := service.classes.Get(ctx, classID)
	if err != nil {
		return err
	}
	if item == nil {
		return classes.ErrNotFound
	}
	if item.Status != "current" {
		return ErrClassReadOnly
	}
	return nil
}

func (service *Service) record(ctx context.Context, action, id, summary string) {
	if service.audit != nil {
		_ = service.audit.Record(ctx, audit.Entry{
			Action: action, EntityType: "custom_field", EntityID: id, Summary: summary,
		})
	}
}

func supportedType(value FieldType) bool {
	return slices.Contains([]FieldType{
		TypeText, TypeLongText, TypeNumber, TypeDate,
		TypeSingleChoice, TypeMultiChoice, TypeBoolean,
	}, value)
}

func normalizeOptions(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}

func validateValue(field FieldDefinition, value any) string {
	switch field.Type {
	case TypeText, TypeLongText:
		if _, ok := value.(string); !ok {
			return "请输入文本"
		}
	case TypeNumber:
		switch value.(type) {
		case int, int32, int64, float32, float64:
		default:
			return "请输入数字"
		}
	case TypeDate:
		text, ok := value.(string)
		if !ok {
			return "日期格式应为 YYYY-MM-DD"
		}
		if _, err := time.Parse("2006-01-02", text); err != nil {
			return "日期格式应为 YYYY-MM-DD"
		}
	case TypeSingleChoice:
		text, ok := value.(string)
		if !ok || !slices.Contains(field.Options, text) {
			return "选项不在允许范围内"
		}
	case TypeMultiChoice:
		values, ok := value.([]string)
		if !ok {
			return "请选择一个或多个选项"
		}
		for _, selected := range values {
			if !slices.Contains(field.Options, selected) {
				return "选项不在允许范围内"
			}
		}
	case TypeBoolean:
		if _, ok := value.(bool); !ok {
			return "请选择是或否"
		}
	}
	return ""
}
