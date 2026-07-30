package webassets_test

import (
	"bytes"
	"io/fs"
	"testing"

	webassets "github.com/kejixiaoliang/ClassPilot/web"
)

func TestEmbeddedIndexExists(t *testing.T) {
	body, err := fs.ReadFile(webassets.Dist, "dist/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("ClassPilot")) {
		t.Fatal("内嵌首页缺少产品标识")
	}
}
