package classes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
)

type Service struct {
	repository *Repository
	audit      *audit.Service
	now        func() time.Time
}

func NewService(repository *Repository, auditService *audit.Service) *Service {
	return &Service{
		repository: repository,
		audit:      auditService,
		now:        time.Now,
	}
}

func (service *Service) CreateCurrent(ctx context.Context, input CreateInput) (Class, error) {
	if err := validateCreate(input); err != nil {
		return Class{}, err
	}
	current, err := service.repository.Current(ctx)
	if err != nil {
		return Class{}, err
	}
	if current != nil {
		return Class{}, ErrCurrentExists
	}

	now := service.now().UTC()
	item := Class{
		ID:         uuid.NewString(),
		Name:       strings.TrimSpace(input.Name),
		Stage:      strings.TrimSpace(input.Stage),
		Grade:      strings.TrimSpace(input.Grade),
		SchoolYear: strings.TrimSpace(input.SchoolYear),
		Status:     "current",
		Notes:      strings.TrimSpace(input.Notes),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err = service.repository.Insert(ctx, item); err != nil {
		if strings.Contains(err.Error(), "one_current_class") {
			return Class{}, ErrCurrentExists
		}
		return Class{}, err
	}
	if service.audit != nil {
		_ = service.audit.Record(ctx, audit.Entry{
			Action: "class.created", EntityType: "class", EntityID: item.ID, Summary: "创建当前班级",
		})
	}
	return item, nil
}

func (service *Service) GetCurrent(ctx context.Context) (*Class, error) {
	return service.repository.Current(ctx)
}

func (service *Service) Get(ctx context.Context, id string) (*Class, error) {
	item, err := service.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (service *Service) Update(ctx context.Context, id string, input UpdateInput) (Class, error) {
	item, err := service.Get(ctx, id)
	if err != nil {
		return Class{}, err
	}
	if item.Status == "archived" {
		return Class{}, ErrArchivedReadOnly
	}
	if strings.TrimSpace(input.Name) != "" {
		item.Name = strings.TrimSpace(input.Name)
	}
	if strings.TrimSpace(input.Stage) != "" {
		item.Stage = strings.TrimSpace(input.Stage)
	}
	if strings.TrimSpace(input.Grade) != "" {
		item.Grade = strings.TrimSpace(input.Grade)
	}
	if strings.TrimSpace(input.SchoolYear) != "" {
		item.SchoolYear = strings.TrimSpace(input.SchoolYear)
	}
	item.Notes = strings.TrimSpace(input.Notes)
	item.UpdatedAt = service.now().UTC()

	if err = service.repository.Update(ctx, *item); err != nil {
		return Class{}, err
	}
	if service.audit != nil {
		_ = service.audit.Record(ctx, audit.Entry{
			Action: "class.updated", EntityType: "class", EntityID: item.ID, Summary: "更新班级资料",
		})
	}
	return *item, nil
}

func (service *Service) ListArchived(ctx context.Context) ([]Class, error) {
	return service.repository.ListArchived(ctx)
}

func (service *Service) ArchiveCurrent(ctx context.Context, id string) error {
	item, err := service.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Status != "current" {
		return ErrArchivedReadOnly
	}
	if err = service.repository.Archive(ctx, id, service.now().UTC()); err != nil {
		return err
	}
	if service.audit != nil {
		_ = service.audit.Record(ctx, audit.Entry{
			Action: "class.archived", EntityType: "class", EntityID: id, Summary: "归档当前班级",
		})
	}
	return nil
}

func (service *Service) RestoreAsCurrent(ctx context.Context, id string) error {
	current, err := service.repository.Current(ctx)
	if err != nil {
		return err
	}
	if current != nil {
		return ErrCurrentExists
	}
	item, err := service.Get(ctx, id)
	if err != nil {
		return err
	}
	if item.Status != "archived" {
		return errors.New("班级不是归档状态")
	}
	if err = service.repository.Restore(ctx, id, service.now().UTC()); err != nil {
		return err
	}
	if service.audit != nil {
		_ = service.audit.Record(ctx, audit.Entry{
			Action: "class.restored", EntityType: "class", EntityID: id, Summary: "恢复历史班级",
		})
	}
	return nil
}

func validateCreate(input CreateInput) error {
	if strings.TrimSpace(input.Name) == "" ||
		strings.TrimSpace(input.Stage) == "" ||
		strings.TrimSpace(input.Grade) == "" ||
		strings.TrimSpace(input.SchoolYear) == "" {
		return ErrInvalidInput
	}
	return nil
}
