# CutX

> 跨平台离线文件切割与合并工具 — 一个二进制，跨 macOS / Windows / Linux 三平台。
>
> Cross-platform offline file splitter & merger — one binary for macOS, Windows, and Linux.

[English](#english) | [中文](#中文)

---

## 中文

### 下载

从 [Releases](../../releases) 页面下载对应平台的二进制文件：

| 平台 | 文件 | 说明 |
|------|------|------|
| macOS Intel | `cutx-darwin-amd64` | CLI 交互式菜单 |
| macOS Apple Silicon | `cutx-darwin-arm64` | CLI 交互式菜单 |
| Linux x86_64 | `cutx-linux-amd64` | CLI 交互式菜单 |
| Linux ARM64 | `cutx-linux-arm64` | CLI 交互式菜单 |
| Windows | `CutX-windows-amd64.exe` | GUI 图形界面，双击即用 |

### 快速开始

#### Windows

双击 `CutX.exe` 打开图形界面，选择操作（切割/合并/校验），拖拽文件或点击浏览，设置参数后点击开始。

#### macOS / Linux

```bash
chmod +x cutx-darwin-arm64
./cutx-darwin-arm64
```

显示交互式菜单：
```
╔══════════════════════════════════════════════╗
║          CutX — 文件切割与合并工具           ║
╚══════════════════════════════════════════════╝

请选择操作:
  1. 切割文件 (split)
  2. 合并文件 (merge)
  3. 校验分片 (verify)
  4. 显示帮助
  0. 退出
```

### 命令行用法

```bash
# 切割（默认2G，自动创建 cutx-<文件名>/ 子目录）
cutx split ubuntu.img -s 2G

# 自定义大小 + SHA-256
cutx split ubuntu.img -s 500M --hash sha256

# 合并（快速模式）
cutx merge cutx-ubuntu/ubuntu.img.manifest.json

# 合并（边写边校验）
cutx merge cutx-ubuntu/ubuntu.img.manifest.json --mode verify

# 仅校验
cutx verify cutx-ubuntu/ubuntu.img.manifest.json
```

---

## English

### Download

Download the binary for your platform from the [Releases](../../releases) page:

| Platform | File | Description |
|----------|------|-------------|
| macOS Intel | `cutx-darwin-amd64` | CLI interactive menu |
| macOS Apple Silicon | `cutx-darwin-arm64` | CLI interactive menu |
| Linux x86_64 | `cutx-linux-amd64` | CLI interactive menu |
| Linux ARM64 | `cutx-linux-arm64` | CLI interactive menu |
| Windows | `CutX-windows-amd64.exe` | GUI app, double-click to run |

### Quick Start

#### Windows

Double-click `CutX.exe` to open the GUI. Select an operation (Split / Merge / Verify), drag & drop a file or click browse, set parameters, and click Start.

#### macOS / Linux

```bash
chmod +x cutx-darwin-arm64
./cutx-darwin-arm64
```

Interactive menu:
```
╔══════════════════════════════════════════════╗
║          CutX — File Splitter & Merger       ║
╚══════════════════════════════════════════════╝

Select operation:
  1. Split file
  2. Merge file
  3. Verify chunks
  4. Help
  0. Exit
```

### Command Line

```bash
# Split (default 2G, auto-creates cutx-<filename>/ subdirectory)
cutx split ubuntu.img -s 2G

# Custom size + SHA-256
cutx split ubuntu.img -s 500M --hash sha256

# Merge (quick mode)
cutx merge cutx-ubuntu/ubuntu.img.manifest.json

# Merge (verify while writing)
cutx merge cutx-ubuntu/ubuntu.img.manifest.json --mode verify

# Verify only
cutx verify cutx-ubuntu/ubuntu.img.manifest.json
```

---

## 项目结构 / Project Structure

```
cutx/
├── main.go                     # CLI 入口 / CLI entry
├── cmd/                        # CLI 命令 / CLI commands (Cobra)
│   ├── root.go                # 根命令
│   ├── split.go               # split 子命令
│   ├── merge.go               # merge 子命令
│   ├── verify.go              # verify 子命令
│   ├── version.go             # version 子命令
│   ├── interactive.go         # 交互式菜单 / Interactive menu
│   └── reporter.go            # CLI Reporter 实现
├── internal/
│   ├── engine/                # 核心逻辑 / Core logic (Split/Merge/Verify + Reporter interface)
│   ├── chunk/                 # 分片命名 / Chunk naming, size parsing, disk detection
│   ├── manifest/              # 清单文件 / Manifest read/write & validation
│   ├── checksum/              # 校验算法 / MD5/SHA-256 wrappers
│   └── progress/              # 进度条 / Progress bar
├── gui/                        # Windows GUI (Wails + Vue3)
│   ├── main.go                # Wails 入口
│   ├── app.go                 # 后端 API / Backend API (calls engine)
│   ├── wails.json             # Wails 配置
│   ├── build/                 # 图标资源 / Icon assets
│   │   ├── appicon.png        # 应用图标 / App icon
│   │   └── windows/icon.ico   # exe 图标 / exe icon
│   └── frontend/              # Vue3 + TailwindCSS 前端
│       └── src/App.vue        # 主界面 / Main UI
├── assets/                    # Logo 素材 / Logo assets
├── docs/                       # 文档 / Documentation
│   ├── PRD.md                 # 产品需求文档
│   ├── feature_list.md        # 功能清单
│   ├── prototype.html         # 产品原型
│   └── adr/                   # 架构决策记录 / Architecture Decision Records
├── CONTEXT.md                 # 领域模型术语表 / Domain glossary
├── Makefile                   # 构建脚本 / Build script
├── go.mod                     # Go 模块 / Go module
└── .github/workflows/build.yml  # CI 构建 / CI build
```

## 技术栈 / Tech Stack

- **CLI 后端 / Backend**: Go + Cobra
- **Windows GUI**: Wails v2 + Vue3 + TailwindCSS
- **校验 / Checksum**: MD5 (默认) / SHA-256
- **构建 / Build**: GitHub Actions CI

## License

MIT
