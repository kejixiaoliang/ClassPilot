package diagnostics

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	phonePattern       = regexp.MustCompile(`1[3-9]\d{9}`)
	idCardPattern      = regexp.MustCompile(`\d{17}[\dXx]`)
	windowsPathPattern = regexp.MustCompile(`(?i)[a-z]:\\[^\s"]+`)
)

type Logger struct {
	output io.Writer
	mutex  sync.Mutex
}

func New(output io.Writer) *Logger {
	return &Logger{output: output}
}

func (logger *Logger) Error(code string, err error, fields map[string]any) {
	safeFields := make(map[string]any, len(fields))
	for key, value := range fields {
		if isSensitiveKey(key) {
			safeFields[key] = "[已脱敏]"
			continue
		}
		safeFields[key] = redactText(fmt.Sprint(value))
	}

	message := ""
	if err != nil {
		message = redactText(err.Error())
	}
	record := map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339Nano),
		"level":   "error",
		"code":    code,
		"message": message,
		"fields":  safeFields,
	}

	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	_ = json.NewEncoder(logger.output).Encode(record)
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	for _, fragment := range []string{"phone", "mobile", "idcard", "identity", "path", "content"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func redactText(value string) string {
	value = phonePattern.ReplaceAllString(value, "[已脱敏]")
	value = idCardPattern.ReplaceAllString(value, "[已脱敏]")
	return windowsPathPattern.ReplaceAllString(value, "[已脱敏]")
}
