.PHONY: all clean build build-cli build-gui

VERSION ?= v1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X 'github.com/pengyongshi/cutx/cmd.version=$(VERSION)' -X 'github.com/pengyongshi/cutx/cmd.commit=$(COMMIT)'

# Build CLI for current platform
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o cutx .

# Cross-compile CLI for all platforms
build-cli:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/cutx-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/cutx-darwin-arm64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/cutx-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/cutx-linux-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/cutx-windows-amd64.exe .

# Build Windows GUI (requires Wails CLI + Node.js, run on Windows)
build-gui:
	cd gui/frontend && npm install && npm run build
	cd gui && wails build -platform windows/amd64 -nsis false

all: build-cli

clean:
	rm -rf dist/ cutx
