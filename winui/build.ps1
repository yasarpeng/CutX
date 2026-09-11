# CutX.GUI WinUI 3 Build Script
# Run on Windows with: pwsh build.ps1

param(
    [string]$Configuration = "Release",
    [string]$Arch = "x64"
)

Write-Host "=== CutX.GUI WinUI 3 Build ===" -ForegroundColor Cyan
Write-Host "Configuration: $Configuration"
Write-Host "Architecture: $Arch"
Write-Host ""

# Check prerequisites
$dotnetVersion = dotnet --version 2>$null
if (-not $dotnetVersion) {
    Write-Host "✗ .NET SDK not found. Install .NET SDK 10+ first." -ForegroundColor Red
    exit 1
}
Write-Host "✓ .NET SDK: $dotnetVersion"

# Check Developer Mode
$devMode = Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock" -Name "AllowDevelopmentWithoutDevLicense" -ErrorAction SilentlyContinue
if ($devMode.AllowDevelopmentWithoutDevLicense -eq 1) {
    Write-Host "✓ Developer Mode: Enabled"
} else {
    Write-Host "⚠ Developer Mode: Not enabled (Settings → System → For developers → On)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Building CutX.GUI..." -ForegroundColor Cyan

# Restore and build
dotnet build CutX.GUI/CutX.GUI.csproj -c $Configuration -p:EnableWindowsTargeting=true -p:Platform=$Arch 2>&1 | ForEach-Object {
    if ($_ -match "error") { Write-Host $_ -ForegroundColor Red }
    elseif ($_ -match "warning") { Write-Host $_ -ForegroundColor Yellow }
    else { Write-Host $_ }
}

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Build succeeded!" -ForegroundColor Green
    
    # Find output
    $outputPath = "CutX.GUI/bin/$Configuration/net10.0-windows10.0.26100.0/win-$Arch"
    if (Test-Path $outputPath) {
        $exePath = Get-ChildItem "$outputPath/CutX.GUI.exe" -ErrorAction SilentlyContinue
        if ($exePath) {
            Write-Host "Output: $($exePath.FullName)" -ForegroundColor Green
            Write-Host "Size: $([math]::Round($exePath.Length / 1MB, 1)) MB"
        }
    }
    
    Write-Host ""
    Write-Host "To run: dotnet run --project CutX.GUI" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "✗ Build failed. See errors above." -ForegroundColor Red
    exit 1
}
