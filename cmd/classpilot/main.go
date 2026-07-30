package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kejixiaoliang/ClassPilot/internal/app"
	"github.com/kejixiaoliang/ClassPilot/internal/server"
	webassets "github.com/kejixiaoliang/ClassPilot/web"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if err := run(); err != nil {
		log.Printf("ClassPilot 启动失败：%v", err)
		os.Exit(1)
	}
}

func run() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径：%w", err)
	}
	defaultWorkspace := filepath.Dir(executable)
	workspaceRoot := flag.String("workspace", defaultWorkspace, "工作台数据目录")
	noBrowser := flag.Bool("no-browser", false, "不自动打开浏览器")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime, err := app.Prepare(ctx, *workspaceRoot, app.BuildInfo{
		Version: version,
		Commit:  commit,
	})
	if err != nil {
		return err
	}
	defer runtime.Close()

	token, err := server.NewToken()
	if err != nil {
		return fmt.Errorf("生成临时访问令牌：%w", err)
	}
	listener, err := server.ListenLoopback()
	if err != nil {
		return fmt.Errorf("启动本地监听：%w", err)
	}

	dist, err := fs.Sub(webassets.Dist, "dist")
	if err != nil {
		listener.Close()
		return fmt.Errorf("读取 Web 资源：%w", err)
	}
	handler := server.New(server.Config{
		Token:    token,
		ReadOnly: runtime.ReadOnly,
		Static:   http.FileServer(http.FS(dist)),
	})
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- httpServer.Serve(listener)
	}()

	url := fmt.Sprintf("http://%s/#token=%s", listener.Addr().String(), token)
	fmt.Printf("ClassPilot %s 已启动：%s\n", version, url)
	if !*noBrowser {
		if err = openBrowser(url); err != nil {
			log.Printf("无法自动打开浏览器，请手动访问上方地址：%v", err)
		}
	}

	select {
	case <-ctx.Done():
	case err = <-serveErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("本地服务异常退出：%w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}
