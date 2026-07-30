package customfields

import (
	"errors"
	"time"
)

type FieldType string

const (
	TypeText         FieldType = "text"
	TypeLongText     FieldType = "long_text"
	TypeNumber       FieldType = "number"
	TypeDate         FieldType = "date"
	TypeSingleChoice FieldType = "single_choice"
	TypeMultiChoice  FieldType = "multi_choice"
	TypeBoolean      FieldType = "boolean"
)

var (
	ErrUnsupportedType = errors.New("不支持的自定义字段类型")
	ErrInvalidInput    = errors.New("自定义字段信息不完整")
	ErrClassReadOnly   = errors.New("历史班级为只读状态")
	ErrNotFound        = errors.New("自定义字段不存在")
)

type FieldDefinition struct {
	ID          string    `json:"id"`
	ClassID     string    `json:"classId"`
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Description string    `json:"description"`
	Required    bool      `json:"required"`
	Enabled     bool      `json:"enabled"`
	ShowInList  bool      `json:"showInList"`
	Options     []string  `json:"options"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ClassID     string    `json:"classId"`
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Description string    `json:"description"`
	Required    bool      `json:"required"`
	ShowInList  bool      `json:"showInList"`
	Options     []string  `json:"options"`
}

type UpdateInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	ShowInList  bool     `json:"showInList"`
	Options     []string `json:"options"`
}
