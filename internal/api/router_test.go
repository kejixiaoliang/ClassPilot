package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/api"
	"github.com/kejixiaoliang/ClassPilot/internal/database"
	"github.com/kejixiaoliang/ClassPilot/internal/server"
)

func TestClassStudentAndCustomFieldAPIFlow(t *testing.T) {
	handler := newTestHandler(t)

	class := requestJSON(t, handler, http.MethodPost, "/api/v1/classes/current", map[string]any{
		"name": "七年级一班", "stage": "初中", "grade": "七年级", "schoolYear": "2026-2027",
	}, http.StatusCreated)
	classID := stringValue(t, class, "id")

	student := requestJSON(t, handler, http.MethodPost, "/api/v1/students", map[string]any{
		"classId": classID, "name": "张三", "studentNumber": "2026001",
	}, http.StatusCreated)
	if stringValue(t, student, "name") != "张三" {
		t.Fatalf("学生响应=%v", student)
	}

	page := requestJSON(t, handler, http.MethodGet, "/api/v1/students?classId="+classID, nil, http.StatusOK)
	if int(page["total"].(float64)) != 1 {
		t.Fatalf("学生列表=%v", page)
	}

	field := requestJSON(t, handler, http.MethodPost, "/api/v1/custom-fields", map[string]any{
		"classId": classID, "name": "住宿情况", "type": "single_choice", "options": []string{"走读", "住校"},
	}, http.StatusCreated)
	if stringValue(t, field, "name") != "住宿情况" {
		t.Fatalf("字段响应=%v", field)
	}
}

func TestAPIMapsDuplicateStudentNumberToConflict(t *testing.T) {
	handler := newTestHandler(t)
	class := requestJSON(t, handler, http.MethodPost, "/api/v1/classes/current", map[string]any{
		"name": "一班", "stage": "初中", "grade": "七年级", "schoolYear": "2026-2027",
	}, http.StatusCreated)
	classID := stringValue(t, class, "id")
	body := map[string]any{"classId": classID, "name": "张三", "studentNumber": "001"}
	requestJSON(t, handler, http.MethodPost, "/api/v1/students", body, http.StatusCreated)
	body["name"] = "李四"

	errorBody := requestError(t, handler, http.MethodPost, "/api/v1/students", body, http.StatusConflict)
	if stringValue(t, errorBody, "code") != "STUDENT_NUMBER_EXISTS" {
		t.Fatalf("错误响应=%v", errorBody)
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "api.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return server.New(server.Config{Token: "secret", API: api.New(db)})
}

func requestJSON(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) map[string]any {
	t.Helper()
	payload := request(t, handler, method, path, body, wantStatus)
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("响应缺少 data：%v", payload)
	}
	return data
}

func requestError(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) map[string]any {
	t.Helper()
	payload := request(t, handler, method, path, body, wantStatus)
	data, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("响应缺少 error：%v", payload)
	}
	return data
}

func request(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) map[string]any {
	t.Helper()
	var encoded []byte
	if body != nil {
		var err error
		encoded, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	request.Header.Set("X-ClassPilot-Token", "secret")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != wantStatus {
		t.Fatalf("%s %s 状态码=%d，期望=%d，响应=%s", method, path, response.Code, wantStatus, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func stringValue(t *testing.T, values map[string]any, key string) string {
	t.Helper()
	value, ok := values[key].(string)
	if !ok {
		t.Fatalf("%s 不是字符串：%v", key, values[key])
	}
	return value
}
