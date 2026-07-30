package server_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/server"
)

func TestAPIRejectsMissingToken(t *testing.T) {
	handler := server.New(server.Config{Token: "secret"})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("状态码=%d，期望=%d", response.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, response, "UNAUTHORIZED")
}

func TestAPIAcceptsValidToken(t *testing.T) {
	handler := server.New(server.Config{Token: "secret"})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	request.Header.Set("X-ClassPilot-Token", "secret")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("状态码=%d，期望=%d", response.Code, http.StatusOK)
	}
}

func TestReadOnlyModeRejectsWriteRequests(t *testing.T) {
	handler := server.New(server.Config{Token: "secret", ReadOnly: true})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/classes/current", nil)
	request.Header.Set("X-ClassPilot-Token", "secret")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusLocked {
		t.Fatalf("状态码=%d，期望=%d", response.Code, http.StatusLocked)
	}
	assertErrorCode(t, response, "DATABASE_READ_ONLY")
}

func TestListenLoopbackNeverBindsLAN(t *testing.T) {
	listener, err := server.ListenLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	host, _, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("监听地址=%s，期望=127.0.0.1", host)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != want {
		t.Fatalf("错误码=%q，期望=%q", body.Error.Code, want)
	}
}
