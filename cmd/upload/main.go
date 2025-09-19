package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"upload-util/pkg/upload"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		filePath   = flag.String("file", "", "要上传的文件路径")
		operation  = flag.String("op", "upload", "操作类型: upload, delete, geturl")
		key        = flag.String("key", "", "文件键名（用于删除和获取URL）")
		verbose    = flag.Bool("v", false, "详细输出")
		version    = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Upload Util CLI\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}

	// 根据操作类型验证参数
	switch *operation {
	case "upload":
		if *filePath == "" {
			printUsage()
			os.Exit(1)
		}
	case "delete", "geturl":
		if *key == "" {
			fmt.Printf("❌ %s 操作需要指定文件键名\n", *operation)
			printUsage()
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 不支持的操作: %s\n", *operation)
		printUsage()
		os.Exit(1)
	}

	// 加载配置
	cfg, err := upload.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 创建上传器
	uploader, err := upload.NewUploader(cfg)
	if err != nil {
		log.Fatalf("❌ 创建上传器失败: %v", err)
	}

	ctx := context.Background()

	switch *operation {
	case "upload":
		result, err := uploadFile(ctx, uploader, *filePath, *verbose)
		if err != nil {
			log.Fatalf("❌ 上传失败: %v", err)
		}
		printUploadResult(result, *verbose)

	case "delete":
		err := uploader.Delete(ctx, *key)
		if err != nil {
			log.Fatalf("❌ 删除失败: %v", err)
		}
		fmt.Printf("✅ 删除成功: %s\n", *key)

	case "geturl":
		url, err := uploader.GetURL(ctx, *key)
		if err != nil {
			log.Fatalf("❌ 获取URL失败: %v", err)
		}
		if *verbose {
			fmt.Printf("🔗 文件键名: %s\n", *key)
			fmt.Printf("🔗 访问URL: %s\n", url)
		} else {
			fmt.Printf("%s\n", url)
		}
	}
}

func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  上传文件:")
	fmt.Println("    upload-cli -file=/path/to/file.jpg")
	fmt.Println("    upload-cli -file=./image.png -v")
	fmt.Println("")
	fmt.Println("  删除文件:")
	fmt.Println("    upload-cli -op=delete -key=uploads/abc123.jpg")
	fmt.Println("")
	fmt.Println("  获取URL:")
	fmt.Println("    upload-cli -op=geturl -key=uploads/abc123.jpg")
	fmt.Println("    upload-cli -op=geturl -key=uploads/abc123.jpg -v")
	fmt.Println("")
	fmt.Println("  其他选项:")
	fmt.Println("    -config=path/to/config.yaml  指定配置文件")
	fmt.Println("    -v                           详细输出")
	fmt.Println("    -version                     显示版本信息")
}

func uploadFile(ctx context.Context, uploader upload.Uploader, filePath string, verbose bool) (*upload.UploadResult, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filePath)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 创建文件头
	header := &multipart.FileHeader{
		Filename: filepath.Base(filePath),
		Size:     stat.Size(),
	}

	if verbose {
		fmt.Printf("⏳ 正在上传: %s (%s)\n", header.Filename, formatFileSize(header.Size))
	}

	// 上传文件
	result, err := uploader.Upload(ctx, file, header)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func printUploadResult(result *upload.UploadResult, verbose bool) {
	fmt.Printf("✅ 上传成功!\n")
	fmt.Printf("🔗 URL: %s\n", result.URL)

	if verbose {
		fmt.Printf("🔑 Key: %s\n", result.Key)
		fmt.Printf("📏 Size: %s\n", formatFileSize(result.Size))
		fmt.Printf("📄 Type: %s\n", result.MimeType)
	} else {
		fmt.Printf("🔑 Key: %s\n", result.Key)
	}
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
