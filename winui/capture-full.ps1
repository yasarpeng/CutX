# Capture full diagnostic build output
$logFile = "build-full-log.txt"
Write-Host "Running diagnostic build... output to $logFile"

dotnet build CutX.GUI -c Release -p:EnableWindowsTargeting=true -p:Platform=x64 -v:diag > $logFile 2>&1

Write-Host ""
Write-Host "=== Error lines ==="
Get-Content $logFile | Select-String "error|Error|exception|Exception|fail|Fail|XamlCompiler" | Select-Object -First 30

Write-Host ""
Write-Host "=== Lines mentioning input.json or output.json ==="
Get-Content $logFile | Select-String "input\.json|output\.json" | Select-Object -First 10

Write-Host ""
Write-Host "=== Last 20 lines of build log ==="
Get-Content $logFile | Select-Object -Last 20
