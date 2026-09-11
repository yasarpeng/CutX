# 从 Walk/Wails 迁移到 Wails v2 + Vue3

CutX 的 Windows GUI 从 lxn/walk（原生 Win32）迁移到 Wails v2（Go + WebView2 + Vue3）。GUI 代码放在独立子目录 `gui/`，通过 `replace` 指令引用父 module 的 `internal/engine`。

**为什么选 Wails**：Walk 的控件样式陈旧，自定义深色主题成本高；WinUI 3 需要 .NET 运行时且 XamlCompiler 在 CI 上频繁崩溃。Wails 用 WebView2（Win10 1903+ 自带），前端用 Vue3 + TailwindCSS，能实现现代 UI 风格。单 exe 产出，绿色免安装。

**为什么独立 gui/ 子目录**：Wails CLI 需要标准的 `wails.json` + `frontend/` 结构，放在子目录不影响父 module 的 CLI 逻辑。`gui/go.mod` 用 `replace github.com/pengyongshi/cutx => ../` 引用 engine 包。

**为什么 Vue3 而非 Svelte**：用户指定 Vue3，SFC 写法直观，中文生态好。40KB 运行时在此场景下无所谓。

**考虑过的替代方案**：
- Walk（保持现状）：被否决，控件样式无法满足现代设计要求。
- WinUI 3：被否决，XamlCompiler.exe 在 GitHub Actions 上崩溃，构建不稳定。
- 纯 Web 服务器 + 浏览器：被否决，用户要求原生窗口体验。
