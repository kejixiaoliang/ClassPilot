package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/kejixiaoliang/ClassPilot/internal/audit"
	"github.com/kejixiaoliang/ClassPilot/internal/classes"
	"github.com/kejixiaoliang/ClassPilot/internal/customfields"
	"github.com/kejixiaoliang/ClassPilot/internal/students"
)

type Router struct {
	classes      *classes.Service
	students     *students.Service
	customFields *customfields.Service
	mux          *http.ServeMux
}

func New(db *sql.DB) http.Handler {
	auditService := audit.New(db)
	classRepository := classes.NewRepository(db)
	router := &Router{
		classes:      classes.NewService(classRepository, auditService),
		students:     students.NewService(students.NewRepository(db), classRepository, auditService),
		customFields: customfields.NewService(customfields.NewRepository(db), classRepository, auditService),
		mux:          http.NewServeMux(),
	}
	router.register()
	return router.mux
}

func (router *Router) register() {
	router.mux.HandleFunc("GET /api/v1/classes/current", router.getCurrentClass)
	router.mux.HandleFunc("POST /api/v1/classes/current", router.createCurrentClass)
	router.mux.HandleFunc("PATCH /api/v1/classes/current", router.updateCurrentClass)
	router.mux.HandleFunc("GET /api/v1/classes/archived", router.listArchivedClasses)
	router.mux.HandleFunc("POST /api/v1/classes/{id}/archive", router.archiveClass)
	router.mux.HandleFunc("POST /api/v1/classes/{id}/restore", router.restoreClass)

	router.mux.HandleFunc("GET /api/v1/students", router.listStudents)
	router.mux.HandleFunc("POST /api/v1/students", router.createStudent)
	router.mux.HandleFunc("GET /api/v1/students/{id}", router.getStudent)
	router.mux.HandleFunc("PATCH /api/v1/students/{id}", router.updateStudent)
	router.mux.HandleFunc("POST /api/v1/students/{id}/trash", router.trashStudent)
	router.mux.HandleFunc("POST /api/v1/students/{id}/restore", router.restoreStudent)

	router.mux.HandleFunc("GET /api/v1/custom-fields", router.listCustomFields)
	router.mux.HandleFunc("POST /api/v1/custom-fields", router.createCustomField)
	router.mux.HandleFunc("PATCH /api/v1/custom-fields/{id}", router.updateCustomField)
	router.mux.HandleFunc("POST /api/v1/custom-fields/{id}/disable", router.disableCustomField)
}

func (router *Router) getCurrentClass(response http.ResponseWriter, request *http.Request) {
	item, err := router.classes.GetCurrent(request.Context())
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, item)
}

func (router *Router) createCurrentClass(response http.ResponseWriter, request *http.Request) {
	var input classes.CreateInput
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.classes.CreateCurrent(request.Context(), input)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusCreated, item)
}

func (router *Router) updateCurrentClass(response http.ResponseWriter, request *http.Request) {
	current, err := router.classes.GetCurrent(request.Context())
	if err != nil {
		writeDomainError(response, err)
		return
	}
	if current == nil {
		writeFailure(response, http.StatusNotFound, "CLASS_NOT_FOUND", "当前没有可编辑的班级")
		return
	}
	var input classes.UpdateInput
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.classes.Update(request.Context(), current.ID, input)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, item)
}

func (router *Router) listArchivedClasses(response http.ResponseWriter, request *http.Request) {
	items, err := router.classes.ListArchived(request.Context())
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"items": items})
}

func (router *Router) archiveClass(response http.ResponseWriter, request *http.Request) {
	if err := router.classes.ArchiveCurrent(request.Context(), request.PathValue("id")); err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"archived": true})
}

func (router *Router) restoreClass(response http.ResponseWriter, request *http.Request) {
	if err := router.classes.RestoreAsCurrent(request.Context(), request.PathValue("id")); err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"restored": true})
}

func (router *Router) listStudents(response http.ResponseWriter, request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(request.URL.Query().Get("pageSize"))
	result, err := router.students.List(request.Context(), students.Query{
		ClassID:        request.URL.Query().Get("classId"),
		Search:         request.URL.Query().Get("search"),
		SortBy:         request.URL.Query().Get("sortBy"),
		SortOrder:      request.URL.Query().Get("sortOrder"),
		Page:           page,
		PageSize:       pageSize,
		IncludeTrashed: request.URL.Query().Get("includeTrashed") == "true",
	})
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, result)
}

func (router *Router) createStudent(response http.ResponseWriter, request *http.Request) {
	var input struct {
		ClassID string `json:"classId"`
		students.CreateInput
	}
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.students.Create(request.Context(), input.ClassID, input.CreateInput)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusCreated, item)
}

func (router *Router) getStudent(response http.ResponseWriter, request *http.Request) {
	item, err := router.students.Get(request.Context(), request.PathValue("id"))
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, item)
}

func (router *Router) updateStudent(response http.ResponseWriter, request *http.Request) {
	var input students.UpdateInput
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.students.Update(request.Context(), request.PathValue("id"), input)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, item)
}

func (router *Router) trashStudent(response http.ResponseWriter, request *http.Request) {
	if err := router.students.Trash(request.Context(), request.PathValue("id")); err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"trashed": true})
}

func (router *Router) restoreStudent(response http.ResponseWriter, request *http.Request) {
	if err := router.students.Restore(request.Context(), request.PathValue("id")); err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"restored": true})
}

func (router *Router) listCustomFields(response http.ResponseWriter, request *http.Request) {
	items, err := router.customFields.List(
		request.Context(),
		request.URL.Query().Get("classId"),
		request.URL.Query().Get("includeDisabled") == "true",
	)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]any{"items": items})
}

func (router *Router) createCustomField(response http.ResponseWriter, request *http.Request) {
	var input customfields.CreateInput
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.customFields.Create(request.Context(), input)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusCreated, item)
}

func (router *Router) updateCustomField(response http.ResponseWriter, request *http.Request) {
	var input customfields.UpdateInput
	if !decodeJSON(response, request, &input) {
		return
	}
	item, err := router.customFields.Update(request.Context(), request.PathValue("id"), input)
	if err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, item)
}

func (router *Router) disableCustomField(response http.ResponseWriter, request *http.Request) {
	if err := router.customFields.Disable(request.Context(), request.PathValue("id")); err != nil {
		writeDomainError(response, err)
		return
	}
	writeData(response, http.StatusOK, map[string]bool{"disabled": true})
}

func decodeJSON(response http.ResponseWriter, request *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeFailure(response, http.StatusBadRequest, "INVALID_JSON", "请求数据格式不正确")
		return false
	}
	return true
}

func writeData(response http.ResponseWriter, status int, data any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]any{"data": data})
}

func writeFailure(response http.ResponseWriter, status int, code, message string) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeDomainError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, classes.ErrCurrentExists):
		writeFailure(response, http.StatusConflict, "CURRENT_CLASS_EXISTS", "当前班级已存在")
	case errors.Is(err, classes.ErrArchivedReadOnly), errors.Is(err, students.ErrClassReadOnly), errors.Is(err, customfields.ErrClassReadOnly):
		writeFailure(response, http.StatusConflict, "CLASS_READ_ONLY", "历史班级为只读状态")
	case errors.Is(err, classes.ErrNotFound), errors.Is(err, students.ErrNotFound), errors.Is(err, customfields.ErrNotFound):
		writeFailure(response, http.StatusNotFound, "NOT_FOUND", "请求的数据不存在")
	case errors.Is(err, students.ErrStudentNumberExists):
		writeFailure(response, http.StatusConflict, "STUDENT_NUMBER_EXISTS", "同一班级内学号已存在")
	case errors.Is(err, classes.ErrInvalidInput), errors.Is(err, students.ErrInvalidInput),
		errors.Is(err, customfields.ErrInvalidInput), errors.Is(err, customfields.ErrUnsupportedType):
		writeFailure(response, http.StatusUnprocessableEntity, "VALIDATION_FAILED", strings.TrimSpace(err.Error()))
	default:
		writeFailure(response, http.StatusInternalServerError, "INTERNAL_ERROR", "操作失败，请查看本地日志")
	}
}
