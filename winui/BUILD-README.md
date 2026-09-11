# CutX GUI 构建说明

## 前置条件

1. **.NET SDK 9.0+**（推荐 9.0.100+）
   ```powershell
   winget install Microsoft.DotNet.SDK.9
   ```
   或同时安装 .NET 10：
   ```powershell
   winget install Microsoft.DotNet.SDK.10
   ```

2. **Developer Mode** 开启
   - 设置 → 系统 → 开发者选项 → 开发人员模式 → 开

3. **Visual C++ 运行时**（XamlCompiler.exe 依赖）
   ```powershell
   winget install Microsoft.VCRedist.2017.x64
   ```

## 构建

```powershell
cd winui
dotnet build CutX.GUI -c Release -p:EnableWindowsTargeting=true -p:Platform=x64
```

## 如果 XamlCompiler.exe 仍然失败

尝试以下步骤：

1. 清除 NuGet 缓存：
   ```powershell
   dotnet nuget locals all --clear
   ```

2. 删除 obj 目录后重建：
   ```powershell
   Remove-Item -Recurse -Force CutX.GUI\obj -ErrorAction SilentlyContinue
   dotnet build CutX.GUI -c Release -p:EnableWindowsTargeting=true -p:Platform=x64
   ```

3. 查看详细错误：
   ```powershell
   dotnet build CutX.GUI -c Release -p:EnableWindowsTargeting=true -p:Platform=x64 -v:detailed 2>&1 | Select-String "error " | Select-Object -First 20
   ```

4. 如果提示缺少 .NET 9 运行时：
   ```powershell
   winget install Microsoft.DotNet.Runtime.9
   ```
