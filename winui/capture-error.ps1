# Capture ALL build output to a file
dotnet build CutX.GUI -c Release -p:EnableWindowsTargeting=true -p:Platform=x64 -v:diag > build-full.txt 2>&1
# Show just the error lines
Get-Content build-full.txt | Select-String "error|Error|exception|Exception|XamlCompiler" | Select-Object -First 30
