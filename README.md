# CutX

> 跨平台离线文件切割与合并工具 — 一个二进制，跨 macOS / Windows / Linux 三平台。

## 下载

从 [Releases](../../releases) 页面下载对应平台的二进制文件：

| 平台 | 文件 | 说明 |
|------|------|------|
| macOS Intel | `cutx-darwin-amd64` | CLI 交互式菜单 |
| macOS Apple Silicon | `cutx-darwin-arm64` | CLI 交互式菜单 |
| Linux x86_64 | `cutx-linux-amd64` | CLI 交互式菜单 |
| Linux ARM64 | `cutx-linux-arm64` | CLI 交互式菜单 |
| Windows | `CutX-windows-amd64.exe` | Wails GUI 图形界面 |

## 快速开始

### Windows

双击 `CutX.exe` 打开图形界面，选择操作（切割/合并/校验），拖拽文件或点击浏览，设置参数后点击开始。

### macOS / Linux

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

## 命令行用法

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

## 项目结构

```
cutx/
├── main.go                     # CLI 入口
├── cmd/                        # CLI 命令 (Cobra)
│   ├── root.go                # 根命令
│   ├── split.go               # split 子命令
│   ├── merge.go               # merge 子命令
│   ├── verify.go              # verify 子命令
│   ├── version.go             # version 子命令
│   ├── interactive.go         # 交互式菜单 (macOS/Linux)
│   └── reporter.go            # CLI Reporter 实现
├── internal/
│   ├── engine/                # 核心逻辑 (Split/Merge/Verify + Reporter 接口)
│   ├── chunk/                 # 分片命名、大小解析、磁盘检测
│   ├── manifest/              # 清单文件读写与校验
│   ├── checksum/              # MD5/SHA-256 算法封装
│   └── progress/              # 进度条
├── gui/                        # Windows GUI (Wails + Vue3)
│   ├── main.go                # Wails 入口
│   ├── app.go                 # 后端 API (调用 engine)
│   ├── wails.json             # Wails 配置
│   ├── build/                 # 图标资源
│   │   ├── appicon.png        # 应用图标 (Wails 嵌入)
│   │   └── windows/icon.ico   # Windows exe 图标
│   └── frontend/              # Vue3 + TailwindCSS 前端
│       └── src/App.vue        # 主界面
├── assets/                    # Logo 素材
├── docs/                       # 文档
│   ├── PRD.md                 # 产品需求文档
│   ├── feature_list.md        # 功能清单
│   ├── prototype.html         # 产品原型
│   └── adr/                   # 架构决策记录
├── CONTEXT.md                 # 领域模型术语表
├── Makefile                   # 构建脚本
├── go.mod                     # Go 模块 (CLI)
└── .github/workflows/build.yml  # CI 构建
```

## 技术栈

- **CLI 后端**: Go + Cobra
- **Windows GUI**: Wails v2 + Vue3 + TailwindCSS
- **校验**: MD5 (默认) / SHA-256
- **构建**: GitHub Actions CI (macOS/Linux 交叉编译, Windows Wails 构建)

## License

MIT
