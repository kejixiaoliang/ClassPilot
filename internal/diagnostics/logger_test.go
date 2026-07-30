package diagnostics_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/diagnostics"
)

func TestLoggerRedactsSensitiveValues(t *testing.T) {
	var output bytes.Buffer
	logger := diagnostics.New(&output)

	logger.Error("SAVE_FAILED", errors.New("write failed"), map[string]any{
		"phone": "13812345678",
		"path":  `C:\Users\teacher\student.pdf`,
		"scope": "student",
	})

	line := output.String()
	for _, forbidden := range []string{"13812345678", `C:\Users\teacher\student.pdf`} {
		if strings.Contains(line, forbidden) {
			t.Fatalf("日志泄露敏感值：%s", line)
		}
	}
	for _, required := range []string{"SAVE_FAILED", "student", "[已脱敏]"} {
		if !strings.Contains(line, required) {
			t.Fatalf("日志缺少 %q：%s", required, line)
		}
	}
}
