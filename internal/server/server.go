package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"strings"
)

type Config struct {
	Token    string
	ReadOnly bool
	API      http.Handler
	Static   http.Handler
}

func New(config Config) http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/health", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, Envelope{
			Data: map[string]any{
				"status":   "ok",
				"readOnly": config.ReadOnly,
			},
		})
	})
	if config.API != nil {
		api.Handle("/api/v1/", config.API)
	}

	root := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/") {
			if !tokenMatches(request.Header.Get("X-ClassPilot-Token"), config.Token) {
				writeError(response, http.StatusUnauthorized, "UNAUTHORIZED", "工作台会话无效，请重新启动")
				return
			}
			if config.ReadOnly && isWriteMethod(request.Method) {
				writeError(response, http.StatusLocked, "DATABASE_READ_ONLY", "数据库处于只读保护模式")
				return
			}
			api.ServeHTTP(response, request)
			return
		}

		if config.Static != nil {
			config.Static.ServeHTTP(response, request)
			return
		}
		http.NotFound(response, request)
	})
	return root
}

func ListenLoopback() (net.Listener, error) {
	return net.Listen("tcp4", "127.0.0.1:0")
}

func NewToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func tokenMatches(got, want string) bool {
	if got == "" || want == "" || len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
