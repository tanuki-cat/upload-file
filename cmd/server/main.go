package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"upload-util/internal/config"
	"upload-util/internal/router"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	// 命令行参数
	var (
		configFile = flag.String("config", "config-example.yaml", "配置文件路径")
		port       = flag.String("port", "8080", "服务端口")
		host       = flag.String("host", "0.0.0.0", "服务地址")
		version    = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Upload Util Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}
	// 加载配置
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 设置路由
	r, err := router.SetupRouter(cfg)
	if err != nil {
		log.Fatalf("初始化路由失败: %v", err)
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", *host, *port),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动成功，监听地址: %s", srv.Addr)
		log.Printf("健康检查: http://%s/api/v1/system/health", srv.Addr)
		log.Printf("上传接口: http://%s/api/v1/upload/file", srv.Addr)
		log.Printf("获取访问url: http://%s/api/v1/upload/url", srv.Addr)
		log.Printf("删除接口: http://%s/api/v1/upload/delete", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 优雅关闭服务器，设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	} else {
		log.Println("服务器已安全关闭")
	}
}
