package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pengyongshi/cutx/internal/chunk"
	"github.com/pengyongshi/cutx/internal/engine"
)

// runInteractive is called when cutx is launched with no arguments (e.g., double-clicked).
// It shows a simple interactive menu for split / merge / verify operations.
func runInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════╗")
	fmt.Println("  ║          CutX — 文件切割与合并工具           ║")
	fmt.Println("  ║     跨平台 · 离线运行 · SHA-256/MD5 校验     ║")
	fmt.Printf("  ║              %s                        ║\n", version)
	fmt.Println("  ╚══════════════════════════════════════════════╝")
	fmt.Println()

	for {
		fmt.Println("  请选择操作:")
		fmt.Println("    1. 切割文件 (split)")
		fmt.Println("    2. 合并文件 (merge)")
		fmt.Println("    3. 校验分片 (verify)")
		fmt.Println("    4. 显示帮助")
		fmt.Println("    0. 退出")
		fmt.Println()
		fmt.Print("  请输入选项 (0-4): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		fmt.Println()

		switch input {
		case "1":
			interactiveSplit(reader)
		case "2":
			interactiveMerge(reader)
		case "3":
			interactiveVerify(reader)
		case "4":
			showInteractiveHelp()
		case "0", "q", "quit", "exit":
			fmt.Println("  再见!")
			return
		default:
			fmt.Println("  ✗ 无效选项，请重新输入")
		}
		fmt.Println()
	}
}

func interactiveSplit(reader *bufio.Reader) {
	// Source file
	fmt.Print("  请输入源文件路径 (或拖拽文件到此): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	// Remove surrounding quotes (from drag-and-drop on Windows)
	input = strings.Trim(input, "\"'")
	if input == "" {
		fmt.Println("  ✗ 未输入文件路径")
		return
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(input)
	if err != nil {
		fmt.Printf("  ✗ 路径解析失败: %v\n", err)
		return
	}

	// Check file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ 文件不存在: %s\n", absPath)
		return
	}

	// Chunk size
	fmt.Println()
	fmt.Println("  切割大小:")
	fmt.Println("    1. 2G (默认)")
	fmt.Println("    2. 1G")
	fmt.Println("    3. 500M")
	fmt.Println("    4. 100M")
	fmt.Println("    5. 自定义")
	fmt.Print("  请选择 (1-5): ")
	sizeInput, _ := reader.ReadString('\n')
	sizeInput = strings.TrimSpace(sizeInput)

	var size string
	switch sizeInput {
	case "1", "":
		size = "2G"
	case "2":
		size = "1G"
	case "3":
		size = "500M"
	case "4":
		size = "100M"
	case "5":
		fmt.Print("  请输入大小 (如 2G, 500M, 100K): ")
		customSize, _ := reader.ReadString('\n')
		size = strings.TrimSpace(customSize)
		if size == "" {
			size = "2G"
		}
	default:
		size = "2G"
	}

	// Hash algorithm
	fmt.Println()
	fmt.Println("  校验算法:")
	fmt.Println("    1. MD5 (快速，默认)")
	fmt.Println("    2. SHA-256 (更安全)")
	fmt.Print("  请选择 (1-2): ")
	hashInput, _ := reader.ReadString('\n')
	hashInput = strings.TrimSpace(hashInput)

	var hashAlgo string
	switch hashInput {
	case "2":
		hashAlgo = "sha256"
	default:
		hashAlgo = "md5"
	}

	// Output directory
	fmt.Println()
	fmt.Print("  输出目录 (留空=自动创建 cutx-<文件名>/ 子目录): ")
	outDir, _ := reader.ReadString('\n')
	outDir = strings.TrimSpace(outDir)
	outDir = strings.Trim(outDir, "\"'")

	// Confirm
	fmt.Println()
	fmt.Printf("  ╭───────────────────────────────────╮\n")
	fmt.Printf("  │ 源文件: %s\n", truncatePath(absPath, 45))
	fmt.Printf("  │ 切割大小: %s\n", size)
	fmt.Printf("  │ 校验算法: %s\n", hashAlgo)
	if outDir != "" {
		fmt.Printf("  │ 输出目录: %s\n", truncatePath(outDir, 45))
	} else {
		baseName := chunk.BaseName(absPath)
		stripped := chunk.StripExtensions(baseName)
		dir := filepath.Dir(absPath)
		fmt.Printf("  │ 输出目录: %s/cutx-%s/\n", truncatePath(dir, 30), stripped)
	}
	fmt.Printf("  ╰───────────────────────────────────╯\n")
	fmt.Print("\n  确认开始切割? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Println("  已取消")
		return
	}

	fmt.Println()
	rep := newCLIReporter(false)
	err = engine.Split(engine.SplitOptions{
		SourcePath: absPath,
		ChunkSize:  size,
		OutputDir:  outDir,
		HashAlgo:   hashAlgo,
		ToolVer:    version + " (" + commit + ")",
	}, rep)
	if err != nil {
		fmt.Printf("\n  ✗ 切割失败: %v\n", err)
	}
}

func interactiveMerge(reader *bufio.Reader) {
	fmt.Print("  请输入清单文件路径 (.manifest.json): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.Trim(input, "\"'")
	if input == "" {
		fmt.Println("  ✗ 未输入文件路径")
		return
	}

	absPath, err := filepath.Abs(input)
	if err != nil {
		fmt.Printf("  ✗ 路径解析失败: %v\n", err)
		return
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ 文件不存在: %s\n", absPath)
		return
	}

	fmt.Println()
	fmt.Println("  合并模式:")
	fmt.Println("    1. quick (快速合并，不校验)")
	fmt.Println("    2. verify (边合并边校验)")
	fmt.Print("  请选择 (1-2, 默认1): ")
	modeInput, _ := reader.ReadString('\n')
	modeInput = strings.TrimSpace(modeInput)

	var mode string
	switch modeInput {
	case "2":
		mode = "verify"
	default:
		mode = "quick"
	}

	fmt.Print("  强制覆盖已存在的文件? (y/n, 默认n): ")
	forceInput, _ := reader.ReadString('\n')
	force := strings.TrimSpace(strings.ToLower(forceInput)) == "y"

	fmt.Println()
	rep := newCLIReporter(false)
	err = engine.Merge(engine.MergeOptions{
		ManifestPath: absPath,
		Mode:         mode,
		Force:        force,
	}, rep)
	if err != nil {
		fmt.Printf("\n  ✗ 合并失败: %v\n", err)
	}
}

func interactiveVerify(reader *bufio.Reader) {
	fmt.Print("  请输入清单文件路径 (.manifest.json): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	input = strings.Trim(input, "\"'")
	if input == "" {
		fmt.Println("  ✗ 未输入文件路径")
		return
	}

	absPath, err := filepath.Abs(input)
	if err != nil {
		fmt.Printf("  ✗ 路径解析失败: %v\n", err)
		return
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		fmt.Printf("  ✗ 文件不存在: %s\n", absPath)
		return
	}

	fmt.Println()
	rep := newCLIReporter(false)
	err = engine.Verify(engine.VerifyOptions{
		ManifestPath: absPath,
	}, rep)
	if err != nil {
		fmt.Printf("\n  ✗ 校验失败: %v\n", err)
	}
}

func showInteractiveHelp() {
	fmt.Println("  CutX — 命令行用法:")
	fmt.Println()
	fmt.Println("    cutx split <源文件> -s <大小> [--hash md5|sha256]")
	fmt.Println("    cutx merge <清单文件> [--mode quick|verify] [--force]")
	fmt.Println("    cutx verify <清单文件>")
	fmt.Println("    cutx version")
	fmt.Println()
	fmt.Println("  示例:")
	fmt.Println("    cutx split ubuntu.img -s 2G")
	fmt.Println("    cutx split ubuntu.img -s 500M --hash sha256")
	fmt.Println("    cutx merge cutx-ubuntu/ubuntu.img.manifest.json --mode verify")
	fmt.Println("    cutx verify cutx-ubuntu/ubuntu.img.manifest.json")
}

// truncatePath shortens a path for display, keeping the filename visible.
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	if len(name) >= maxLen-3 {
		return "..." + name[:maxLen-3]
	}
	dirLen := maxLen - len(name) - 3
	return dir[:dirLen] + "..." + string(filepath.Separator) + name
}

// pauseAndWait prints a message and waits for Enter.
// On Windows, this prevents the console window from closing immediately
// when the program is launched by double-clicking the exe.
func pauseAndWait() {
	fmt.Println()
	fmt.Print("  按回车键退出...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
