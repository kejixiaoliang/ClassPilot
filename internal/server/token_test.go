package server_test

import (
	"testing"

	"github.com/kejixiaoliang/ClassPilot/internal/server"
)

func TestNewTokenProducesDistinctUnpredictableValues(t *testing.T) {
	first, err := server.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := server.NewToken()
	if err != nil {
		t.Fatal(err)
	}

	if len(first) < 40 {
		t.Fatalf("令牌长度不足：%d", len(first))
	}
	if first == second {
		t.Fatal("连续生成的令牌不应相同")
	}
}
