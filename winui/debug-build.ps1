# CutX GUI Debug Build Script
# Runs XamlCompiler.exe directly to see the actual error

$ErrorActionPreference = "Continue"
$projectDir = "CutX.GUI"
$objDir = "$projectDir\obj\x64\Release\net9.0-windows10.0.26100.0"

Write-Host "=== Step 1: Standard build (generates input.json) ===" -ForegroundColor Cyan
dotnet build $projectDir -c Release -p:EnableWindowsTargeting=true -p:Platform=x64 2>&1 | Out-Null
Write-Host "Build attempt completed (errors expected)" -ForegroundColor Yellow

Write-Host ""
Write-Host "=== Step 2: Run XamlCompiler.exe directly ===" -ForegroundColor Cyan

$inputJson = "$objDir\input.json"
$outputJson = "$objDir\output.json"
$xamlCompiler = "$env:USERPROFILE\.nuget\packages\microsoft.windowsappsdk\1.7.250513003\tools\net472\XamlCompiler.exe"

if (Test-Path $inputJson) {
    Write-Host "input.json found: $inputJson" -ForegroundColor Green
    Write-Host "Running XamlCompiler.exe..." -ForegroundColor Yellow
    
    $result = & $xamlCompiler $inputJson $outputJson 2>&1
    
    Write-Host ""
    Write-Host "=== XamlCompiler Output ===" -ForegroundColor Cyan
    $result | ForEach-Object { Write-Host $_ }
    
    Write-Host ""
    Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor $(if ($LASTEXITCODE -eq 0) { 'Green' } else { 'Red' })
    
    # Also check if output.json was created with error details
    if (Test-Path $outputJson) {
        Write-Host ""
        Write-Host "=== output.json content ===" -ForegroundColor Cyan
        Get-Content $outputJson -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "input.json not found at $inputJson" -ForegroundColor Red
    Write-Host "Contents of obj dir:" -ForegroundColor Yellow
    Get-ChildItem -Recurse $projectDir\obj -ErrorAction SilentlyContinue | Select-Object FullName
}

Write-Host ""
Write-Host "=== Step 3: Check .NET Framework version ===" -ForegroundColor Cyan
$netFramework = Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\NET Framework Setup\NDP\v4\Full" -Name "Release" -ErrorAction SilentlyContinue
if ($netFramework) {
    $version = $netFramework.Release
    Write-Host ".NET Framework release: $version" -ForegroundColor Green
    if ($version -ge 528040) {
        Write-Host "  >= 4.8.0 (OK)" -ForegroundColor Green
    } elseif ($version -ge 461808) {
        Write-Host "  >= 4.7.2 (OK)" -ForegroundColor Green
    } else {
        Write-Host "  < 4.7.2 (TOO OLD — install .NET Framework 4.7.2+)" -ForegroundColor Red
    }
} else {
    Write-Host ".NET Framework not found! Install 4.7.2+" -ForegroundColor Red
    Write-Host "  winget install Microsoft.DotNet.Framework.DeveloperPack_4" -ForegroundColor Yellow
}
