package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"upload-util/pkg/upload"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

type UploadTask struct {
	FilePath string
	Result   *upload.UploadResult
	Error    error
	Duration time.Duration
}

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		directory  = flag.String("dir", "", "要上传的目录路径")
		pattern    = flag.String("pattern", "*", "文件匹配模式 (支持 *.jpg, *.png 等)")
		recursive  = flag.Bool("r", false, "递归遍历子目录")
		concurrent = flag.Int("c", 3, "并发上传数量")
		dryRun     = flag.Bool("dry-run", false, "试运行，只显示将要上传的文件")
		verbose    = flag.Bool("v", false, "详细输出")
		version    = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Upload Util Batch Uploader\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}
	if *directory == "" {
		fmt.Println("用法:")
		fmt.Println("  批量上传: batch-upload -dir=./photos -pattern='*.jpg' -c=5")
		fmt.Println("  递归上传: batch-upload -dir=./docs -pattern='*.pdf' -r")
		fmt.Println("  试运行:   batch-upload -dir=./files -dry-run")
		os.Exit(1)
	}

	if _, err := os.Stat(*directory); os.IsNotExist(err) {
		log.Fatalf("目录不存在 %s", *directory)
	}

	files, err := findFiles(*directory, *pattern, *recursive)
	if err != nil {
		log.Fatalf("查找文件失败: %v", err)
	}

	if len(files) == 0 {
		fmt.Printf("📁 目录: %s\n", *directory)
		fmt.Printf("🔍 模式: %s\n", *pattern)
		fmt.Printf("📄 没有找到匹配的文件\n")
		return
	}

	fmt.Printf("📁 目录: %s\n", *directory)
	fmt.Printf("🔍 模式: %s\n", *pattern)
	fmt.Printf("📄 找到 %d 个文件\n", len(files))

	if *dryRun {
		fmt.Println("\n🔍 试运行模式 - 将要上传的文件:")
		for i, file := range files {
			stat, _ := os.Stat(file)
			fmt.Printf("  %d. %s (%s)\n", i+1, file, formatFileSize(stat.Size()))
		}
		return
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

	fmt.Printf("🚀 开始批量上传 (并发: %d)...\n\n", *concurrent)

	// 批量上传
	start := time.Now()
	results := batchUpload(context.Background(), uploader, files, *concurrent, *verbose)
	duration := time.Since(start)

	// 打印结果
	printBatchResults(results, duration)

}

func findFiles(directory, pattern string, recursive bool) ([]string, error) {
	var files []string
	if recursive {
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			match, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return err
			}

			if match {
				files = append(files, path)
			}
			return nil
		})
		return files, err
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			match, err := filepath.Match(pattern, entry.Name())
			if err != nil {
				return nil, err
			}
			if match {
				files = append(files, filepath.Join(directory, entry.Name()))
			}
		}
		return files, nil
	}
}

func batchUpload(ctx context.Context, uploader upload.Uploader, files []string, concurrent int, verbose bool) []UploadTask {
	tasks := make(chan string, len(files))
	results := make(chan UploadTask, len(files))

	for _, file := range files {
		tasks <- file
	}
	close(tasks)
	var wg sync.WaitGroup
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for filePath := range tasks {
				result := uploadSingleFile(ctx, uploader, filePath, verbose, workerID)
				results <- result
			}
		}(i + 1)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []UploadTask
	completed := 0
	total := len(files)

	for result := range results {
		allResults = append(allResults, result)
		completed++

		if !verbose {
			// 显示进度
			fmt.Printf("\r⏳ 进度: %d/%d (%.1f%%)", completed, total, float64(completed)/float64(total)*100)
		}
	}

	if !verbose {
		fmt.Printf("\r") // 清除进度行
	}

	return allResults

}

func uploadSingleFile(ctx context.Context, uploader upload.Uploader, filePath string, verbose bool, workerID int) UploadTask {
	start := time.Now()
	task := UploadTask{FilePath: filePath}

	file, err := os.Open(filePath)
	if err != nil {
		task.Error = err
		task.Duration = time.Since(start)
		return task
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	stat, err := file.Stat()
	if err != nil {
		task.Error = err
		task.Duration = time.Since(start)
		return task
	}

	header := &multipart.FileHeader{
		Filename: filepath.Base(filePath),
		Size:     stat.Size(),
	}

	if verbose {
		fmt.Printf("[Worker %d] ⏳ 上传: %s (%s)\n", workerID, filePath, formatFileSize(stat.Size()))
	}

	result, err := uploader.Upload(ctx, file, header)
	task.Result = result
	task.Error = err
	task.Duration = time.Since(start)

	if verbose {
		if err != nil {
			fmt.Printf("[Worker %d] ❌ 失败: %s - %v\n", workerID, filePath, err)
		} else {
			fmt.Printf("[Worker %d] ✅ 完成: %s -> %s (%.2fs)\n", workerID, filepath.Base(filePath), result.URL, task.Duration.Seconds())
		}
	}

	return task
}

func printBatchResults(results []UploadTask, totalDuration time.Duration) {
	var successCount, failCount int
	var totalSize, uploadedSize int64
	var totalUploadTime time.Duration

	fmt.Println("\n📊 批量上传结果:")
	fmt.Println(strings.Repeat("=", 60))

	for _, task := range results {
		// 获取文件大小
		if stat, err := os.Stat(task.FilePath); err == nil {
			totalSize += stat.Size()
		}

		if task.Error != nil {
			fmt.Printf("❌ %s\n   错误: %v\n", task.FilePath, task.Error)
			failCount++
		} else {
			fmt.Printf("✅ %s\n   URL: %s\n   Key: %s\n",
				filepath.Base(task.FilePath), task.Result.URL, task.Result.Key)
			successCount++
			uploadedSize += task.Result.Size
			totalUploadTime += task.Duration
		}
		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📈 统计信息:\n")
	fmt.Printf("   总文件数: %d\n", len(results))
	fmt.Printf("   成功上传: %d\n", successCount)
	fmt.Printf("   失败数量: %d\n", failCount)
	fmt.Printf("   总大小: %s\n", formatFileSize(totalSize))
	fmt.Printf("   上传大小: %s\n", formatFileSize(uploadedSize))
	fmt.Printf("   总耗时: %.2f 秒\n", totalDuration.Seconds())

	if successCount > 0 {
		avgTime := totalUploadTime.Seconds() / float64(successCount)
		fmt.Printf("   平均耗时: %.2f 秒/文件\n", avgTime)

		if uploadedSize > 0 {
			speed := float64(uploadedSize) / totalDuration.Seconds()
			fmt.Printf("   上传速度: %s/秒\n", formatFileSize(int64(speed)))
		}
	}

	if failCount > 0 {
		fmt.Printf("\n⚠️  成功率: %.1f%%\n", float64(successCount)/float64(len(results))*100)
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
