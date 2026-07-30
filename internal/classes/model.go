package classes

import (
	"errors"
	"time"
)

var (
	ErrCurrentExists    = errors.New("当前班级已存在")
	ErrNotFound         = errors.New("班级不存在")
	ErrArchivedReadOnly = errors.New("历史班级为只读状态")
	ErrInvalidInput     = errors.New("班级信息不完整")
)

type Class struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Stage      string     `json:"stage"`
	Grade      string     `json:"grade"`
	SchoolYear string     `json:"schoolYear"`
	Status     string     `json:"status"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}

type CreateInput struct {
	Name       string `json:"name"`
	Stage      string `json:"stage"`
	Grade      string `json:"grade"`
	SchoolYear string `json:"schoolYear"`
	Notes      string `json:"notes"`
}

type UpdateInput struct {
	Name       string `json:"name"`
	Stage      string `json:"stage"`
	Grade      string `json:"grade"`
	SchoolYear string `json:"schoolYear"`
	Notes      string `json:"notes"`
}
