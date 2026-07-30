package students

import (
	"errors"
	"time"
)

var (
	ErrNotFound            = errors.New("学生不存在")
	ErrInvalidInput        = errors.New("学生姓名不能为空")
	ErrStudentNumberExists = errors.New("同一班级内学号已存在")
	ErrClassReadOnly       = errors.New("历史班级为只读状态")
)

type Student struct {
	ID             string     `json:"id"`
	ClassID        string     `json:"classId"`
	Name           string     `json:"name"`
	Gender         string     `json:"gender"`
	StudentNumber  string     `json:"studentNumber"`
	BirthDate      *time.Time `json:"birthDate,omitempty"`
	EnrollmentDate *time.Time `json:"enrollmentDate,omitempty"`
	Status         string     `json:"status"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type CreateInput struct {
	Name           string         `json:"name"`
	Gender         string         `json:"gender"`
	StudentNumber  string         `json:"studentNumber"`
	BirthDate      *time.Time     `json:"birthDate"`
	EnrollmentDate *time.Time     `json:"enrollmentDate"`
	Status         string         `json:"status"`
	Notes          string         `json:"notes"`
	CustomValues   map[string]any `json:"customValues,omitempty"`
}

type UpdateInput = CreateInput

type Query struct {
	ClassID        string
	Search         string
	SortBy         string
	SortOrder      string
	Page           int
	PageSize       int
	IncludeTrashed bool
}

type Page struct {
	Items    []Student `json:"items"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

func AgeAt(birthDate, at time.Time) int {
	age := at.Year() - birthDate.Year()
	if at.Month() < birthDate.Month() ||
		(at.Month() == birthDate.Month() && at.Day() < birthDate.Day()) {
		age--
	}
	return age
}
