package students

import (
	"context"
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

func (service *Service) Create(ctx context.Context, classID string, input CreateInput) (Student, error) {
	if strings.TrimSpace(input.Name) == "" {
		return Student{}, ErrInvalidInput
	}
	if err := service.ensureWritableClass(ctx, classID); err != nil {
		return Student{}, err
	}
	now := service.now().UTC()
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	item := Student{
		ID: uuid.NewString(), ClassID: classID, Name: strings.TrimSpace(input.Name),
		Gender: strings.TrimSpace(input.Gender), StudentNumber: strings.TrimSpace(input.StudentNumber),
		BirthDate: input.BirthDate, EnrollmentDate: input.EnrollmentDate,
		Status: status, Notes: strings.TrimSpace(input.Notes), CreatedAt: now, UpdatedAt: now,
	}
	if err := service.repository.Insert(ctx, item); err != nil {
		if isStudentNumberConflict(err) {
			return Student{}, ErrStudentNumberExists
		}
		return Student{}, err
	}
	service.record(ctx, "student.created", item.ID, "新增学生记录")
	return item, nil
}

func (service *Service) Get(ctx context.Context, id string) (*Student, error) {
	item, err := service.repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

func (service *Service) Update(ctx context.Context, id string, input UpdateInput) (Student, error) {
	item, err := service.Get(ctx, id)
	if err != nil {
		return Student{}, err
	}
	if err = service.ensureWritableClass(ctx, item.ClassID); err != nil {
		return Student{}, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return Student{}, ErrInvalidInput
	}
	item.Name = strings.TrimSpace(input.Name)
	item.Gender = strings.TrimSpace(input.Gender)
	item.StudentNumber = strings.TrimSpace(input.StudentNumber)
	item.BirthDate = input.BirthDate
	item.EnrollmentDate = input.EnrollmentDate
	if strings.TrimSpace(input.Status) != "" {
		item.Status = strings.TrimSpace(input.Status)
	}
	item.Notes = strings.TrimSpace(input.Notes)
	item.UpdatedAt = service.now().UTC()
	if err = service.repository.Update(ctx, *item); err != nil {
		if isStudentNumberConflict(err) {
			return Student{}, ErrStudentNumberExists
		}
		return Student{}, err
	}
	service.record(ctx, "student.updated", item.ID, "更新学生资料")
	return *item, nil
}

func (service *Service) List(ctx context.Context, query Query) (Page, error) {
	return service.repository.List(ctx, query)
}

func (service *Service) Trash(ctx context.Context, id string) error {
	item, err := service.Get(ctx, id)
	if err != nil {
		return err
	}
	if err = service.ensureWritableClass(ctx, item.ClassID); err != nil {
		return err
	}
	now := service.now().UTC()
	if err = service.repository.SetDeletedAt(ctx, id, &now); err != nil {
		return err
	}
	service.record(ctx, "student.trashed", id, "学生移入回收站")
	return nil
}

func (service *Service) Restore(ctx context.Context, id string) error {
	item, err := service.Get(ctx, id)
	if err != nil {
		return err
	}
	if err = service.ensureWritableClass(ctx, item.ClassID); err != nil {
		return err
	}
	if err = service.repository.SetDeletedAt(ctx, id, nil); err != nil {
		if isStudentNumberConflict(err) {
			return ErrStudentNumberExists
		}
		return err
	}
	service.record(ctx, "student.restored", id, "从回收站恢复学生")
	return nil
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
			Action: action, EntityType: "student", EntityID: id, Summary: summary,
		})
	}
}

func isStudentNumberConflict(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unique_student_number_per_class") ||
		strings.Contains(message, "students.class_id, students.student_number")
}
