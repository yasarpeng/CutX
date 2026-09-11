# CutX GUI Build — Windows Container
# Requires: Docker Desktop on Windows with Windows containers mode
# Run: docker build -t cutx-gui -f Dockerfile . 
# (Switch to Windows containers first: Docker Desktop → Settings → switch)

FROM mcr.microsoft.com/dotnet/sdk:9.0-windowsservercore-ltsc2022

WORKDIR /app

# Copy project files
COPY CutX.GUI/ ./CutX.GUI/

# Build
WORKDIR /app/CutX.GUI
RUN dotnet build CutX.GUI.csproj -c Release -p:EnableWindowsTargeting=true -p:Platform=x64

# Output is at: /app/CutX.GUI/bin/Release/net9.0-windows10.0.26100.0/win-x64/
