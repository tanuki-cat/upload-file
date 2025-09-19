package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"upload-util/pkg/upload"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

type CLI struct {
	uploader upload.Uploader
	scanner  *bufio.Scanner
	history  []string
}

func main() {
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		version    = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *version {
		fmt.Printf("Upload Util Interactive CLI\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
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

	cli := &CLI{
		uploader: uploader,
		scanner:  bufio.NewScanner(os.Stdin),
		history:  make([]string, 0),
	}

	fmt.Printf("🚀 Upload Util 交互式命令行 v%s\n", Version)
	fmt.Println("📝 输入 'help' 查看帮助，输入 'quit' 退出")
	fmt.Printf("📁 当前目录: %s\n", getCurrentDir())
	fmt.Println()

	for {
		fmt.Print("upload-util> ")
		if !cli.scanner.Scan() {
			break
		}
		line := strings.TrimSpace(cli.scanner.Text())
		if line == "" {
			continue
		}

		cli.history = append(cli.history, line)
		args := strings.Fields(line)
		cmd := args[0]
		switch cmd {
		case "help", "h":
			cli.printHelp()
		case "upload", "up":
			cli.handleUpload(args[1:])
		case "delete", "del", "rm":
			cli.handleDelete(args[1:])
		case "geturl", "url":
			cli.handleGetURL(args[1:])
		case "ls", "list":
			cli.handleList(args[1:])
		case "cd":
			cli.handleChangeDir(args[1:])
		case "pwd":
			cli.handlePwd()
		case "history":
			cli.handleHistory()
		case "clear":
			cli.handleClear()
		case "quit", "exit", "q":
			fmt.Println("👋 再见!")
			return
		default:
			fmt.Printf("❌ 未知命令: %s，输入 'help' 查看帮助\n", cmd)
		}

	}
}

func (c *CLI) printHelp() {
	fmt.Println("📚 可用命令:")
	fmt.Println("  文件操作:")
	fmt.Println("    upload|up <file>      上传文件")
	fmt.Println("    delete|del|rm <key>   删除文件")
	fmt.Println("    geturl|url <key>      获取文件URL")
	fmt.Println()
	fmt.Println("  目录操作:")
	fmt.Println("    ls|list [pattern]     列出文件")
	fmt.Println("    cd <dir>              切换目录")
	fmt.Println("    pwd                   显示当前目录")
	fmt.Println()
	fmt.Println("  其他:")
	fmt.Println("    history               显示命令历史")
	fmt.Println("    clear                 清屏")
	fmt.Println("    help|h                显示帮助")
	fmt.Println("    quit|exit|q           退出程序")
}

func (c *CLI) handleUpload(args []string) {
	if len(args) < 1 {
		fmt.Print("用法: upload <文件路径>")
		return
	}

	filePath := args[0]
	if !filepath.IsAbs(filePath) {
		wd, _ := os.Getwd()
		filePath = filepath.Join(wd, filePath)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("文件不存在或无法访问: %v\n", err)
		return
	}
	if stat.IsDir() {
		fmt.Printf("%s 是一个目录，请指定文件路径\n", filePath)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)

	header := &multipart.FileHeader{
		Filename: filepath.Base(filePath),
		Size:     stat.Size(),
	}

	fmt.Printf("⏳ 正在上传 %s (%s)...\n", header.Filename, formatFileSize(header.Size))

	result, err := c.uploader.Upload(context.Background(), file, header)
	if err != nil {
		fmt.Printf("❌ 上传失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 上传成功!\n")
	fmt.Printf("   📄 文件: %s\n", result.Key)
	fmt.Printf("   🔗 URL:  %s\n", result.URL)
	fmt.Printf("   📏 大小: %s\n", formatFileSize(result.Size))
	fmt.Printf("   📝 类型: %s\n", result.MimeType)
}

func (c *CLI) handleDelete(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ 用法: delete <文件键名>")
		return
	}

	key := args[0]
	fmt.Printf("⏳ 正在删除: %s\n", key)

	err := c.uploader.Delete(context.Background(), key)
	if err != nil {
		fmt.Printf("❌ 删除失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 删除成功: %s\n", key)
}

func (c *CLI) handleGetURL(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ 用法: geturl <文件键名>")
		return
	}

	key := args[0]
	url, err := c.uploader.GetURL(context.Background(), key)
	if err != nil {
		fmt.Printf("❌ 获取URL失败: %v\n", err)
		return
	}

	fmt.Printf("🔗 URL: %s\n", url)
}

func (c *CLI) handleList(args []string) {
	pattern := "*"
	if len(args) > 0 {
		pattern = args[0]
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取当前目录失败: %v\n", err)
		return
	}

	matches, err := filepath.Glob(filepath.Join(wd, pattern))
	if err != nil {
		fmt.Printf("❌ 匹配文件失败: %v\n", err)
		return
	}

	fmt.Printf("📁 目录: %s\n", wd)
	if pattern != "*" {
		fmt.Printf("🔍 模式: %s\n", pattern)
	}

	if len(matches) == 0 {
		fmt.Println("📄 没有找到匹配的文件")
		return
	}

	fmt.Println("📋 文件列表:")
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		name := filepath.Base(match)
		if stat.IsDir() {
			fmt.Printf("  📁 %s/\n", name)
		} else {
			fmt.Printf("  📄 %-30s %10s\n", name, formatFileSize(stat.Size()))
		}
	}
}

func (c *CLI) handleChangeDir(args []string) {
	if len(args) < 1 {
		// 切换到用户主目录
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ 获取主目录失败: %v\n", err)
			return
		}
		args = []string{homeDir}
	}

	newDir := args[0]

	// 处理特殊情况
	switch newDir {
	case "~":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("❌ 获取主目录失败: %v\n", err)
			return
		}
		newDir = homeDir
	case "..":
		wd, _ := os.Getwd()
		newDir = filepath.Dir(wd)
	case ".":
		return // 当前目录，不需要改变
	}

	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(newDir) {
		wd, _ := os.Getwd()
		newDir = filepath.Join(wd, newDir)
	}

	err := os.Chdir(newDir)
	if err != nil {
		fmt.Printf("❌ 切换目录失败: %v\n", err)
		return
	}

	fmt.Printf("📁 当前目录: %s\n", newDir)
}

func (c *CLI) handlePwd() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ 获取当前目录失败: %v\n", err)
		return
	}
	fmt.Printf("📁 %s\n", wd)
}

func (c *CLI) handleHistory() {
	fmt.Println("📜 命令历史:")
	for i, cmd := range c.history {
		fmt.Printf("  %3d  %s\n", i+1, cmd)
	}
}

func (c *CLI) handleClear() {
	// 清屏
	fmt.Print("\033[2J\033[H")
	fmt.Printf("🚀 Upload Util 交互式命令行 v%s\n", Version)
	fmt.Printf("📁 当前目录: %s\n", getCurrentDir())
	fmt.Println()
}

func getCurrentDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
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
