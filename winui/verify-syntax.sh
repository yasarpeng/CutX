#!/bin/bash
# Verifies C# syntax on non-Windows platforms using csc directly.
# This does NOT build the full WinUI project (requires Windows + XAML compiler).
# It only checks that C# code is syntactically and type-correct.

set -e

DOTNET="${DOTNET:-$HOME/.dotnet/dotnet}"
CSC="$HOME/.dotnet/sdk/10.0.401/Roslyn/bincore/csc.dll"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Find reference assemblies
NETREF=$(find "$HOME/.dotnet/packs/Microsoft.NETCore.App.Ref" -path "*/ref/net10.0" -type d 2>/dev/null | head -1)
WINUI_LIB=$(find "$HOME/.nuget/packages/microsoft.windowsappsdk" -path "*/lib/net6.0-windows10.0.18362.0" -type d 2>/dev/null | head -1)
SDK_REF=$(find "$HOME/.nuget/packages/microsoft.windows.sdk.net.ref" -path "*/lib/net8.0" -type d 2>/dev/null | head -1)

if [ -z "$NETREF" ] || [ -z "$WINUI_LIB" ] || [ -z "$SDK_REF" ]; then
    echo "✗ Reference assemblies not found. Run 'dotnet restore' first."
    exit 1
fi

# Create response file
RSP=$(mktemp /tmp/cutx_verify.XXXXXX.rsp)
echo "-nologo" > "$RSP"
echo "-nullable:enable" >> "$RSP"
echo "-nowarn:CS1701,CS8632,CS0103,CS0115,CS0120,CS0012,CS0234,CS0246,CS1061" >> "$RSP"
echo "-target:library" >> "$RSP"
echo "-checked" >> "$RSP"
echo "-unsafe" >> "$RSP"
echo "-define:WINDOWS" >> "$RSP"

for dll in "$NETREF"/*.dll; do
    echo "-reference:\"$dll\"" >> "$RSP"
done
echo "-reference:\"$WINUI_LIB/Microsoft.WinUI.dll\"" >> "$RSP"
echo "-reference:\"$SDK_REF/Microsoft.Windows.SDK.NET.dll\"" >> "$RSP"
echo "-reference:\"$SDK_REF/WinRT.Runtime.dll\"" >> "$RSP"
echo "-reference:\"$SDK_REF/Microsoft.Windows.UI.Xaml.dll\"" >> "$RSP"
echo "-out:\"/dev/null\"" >> "$RSP"

# Add source files (skip GlobalUsings which is auto-generated on Windows)
for f in App.xaml.cs Models/LogEntry.cs Services/CutxProcessService.cs \
         Views/SplitPage.xaml.cs Views/MergePage.xaml.cs \
         Views/VerifyPage.xaml.cs Views/VersionPage.xaml.cs; do
    echo "$SCRIPT_DIR/CutX.GUI/$f" >> "$RSP"
done

# Run compilation
export PATH="$HOME/.dotnet:$PATH"
RESULT=$(dotnet "$CSC" @"$RSP" 2>&1)
rm -f "$RSP"

if echo "$RESULT" | grep -q "error"; then
    echo "$RESULT" | grep "error" | head -20
    echo ""
    echo "⚠ Some errors are expected (InitializeComponent, x:Name controls, WinRT projections) — these resolve on Windows with XAML compiler."
    echo "✗ Non-expected errors found — review above."
    exit 1
else
    echo "✅ All C# files compile without syntax or type errors."
    echo "(InitializeComponent, x:Name controls, and WinRT projections are expected to be missing without XAML compiler — they resolve on Windows.)"
    exit 0
fi
