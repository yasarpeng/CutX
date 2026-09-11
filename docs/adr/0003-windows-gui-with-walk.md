# Windows 专用 GUI 版本：WinUI 3 + Windows App SDK

V2.0 计划发布 Windows 专用图形界面版本 `CutX.GUI.exe`，基于 WinUI 3 + Windows App SDK，采用 Fluent Design。GUI 作为 C# shell 调用 `cutx.exe` CLI 二进制，通过 `CutxProcessService` 解析 stderr 输出获取进度和日志。核心逻辑保持 Go 实现，GUI 不重复业务逻辑。

**为什么做 GUI**：Windows 用户（尤其客户运维人员）通常不熟悉命令行。PRD V1.0 明确不做 GUI，但 Windows 用户的实际反馈表明纯 CLI 在该平台阻力大。

**为什么选 WinUI 3**：Windows 原生 UI 框架，支持 Fluent Design（Mica 背景、亚克力效果、圆角控件），视觉品质远超 Win32/Walk。WinUI 3 是微软推荐的 Windows 桌面 UI 框架，长期投资价值高。

**为什么 GUI 调用 CLI 而非直接调用 Go core**：Go 和 C# 的跨语言调用成本高（CGO 互操作或 gRPC 通信），而 cutx.exe 已有完善的 CLI 接口和 stderr 输出。GUI 通过子进程调用 CLI + 解析输出，零互操作成本，且 CLI 更新后 GUI 自动兼容。

**为什么放弃 Walk**：Walk 基于 Win32 API，控件样式陈旧（无法满足"禁止默认 Win32 灰色控件"的设计要求），自定义样式成本高。WinUI 3 原生支持 Fluent Design，控件自带圆角、hover 动效、主题切换。

**考虑过的替代方案**：
- Walk (lxn/walk)：被否决，控件样式无法满足设计标准，自定义 ControlTemplate 成本高。
- Fyne（跨平台 GUI）：被否决，体积 ~20MB，超 10MB 目标，且 GUI 只面向 Windows。
- Wails（WebView2）：被否决，依赖 WebView2 运行时，非真正离线。
- GUI 直接调用 Go core（CGO 互操作）：被否决，跨语言调用复杂度高，维护成本大。
