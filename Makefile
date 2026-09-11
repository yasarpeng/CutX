.PHONY: all clean build darwin-amd64 darwin-arm64 windows-amd64 linux-amd64 linux-arm64 icons

VERSION ?= v1.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%d")
LDFLAGS := -s -w -X 'github.com/pengyongshi/cutx/cmd.version=$(VERSION)' -X 'github.com/pengyongshi/cutx/cmd.commit=$(COMMIT)'

BUILDDIR := dist

all: darwin-amd64 darwin-arm64 windows-amd64 linux-amd64 linux-arm64

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

# Generate Windows icon resource (.syso) from .ico
icons:
	@which rsrc > /dev/null 2>&1 || go install github.com/akavel/rsrc@latest
	rsrc -ico build-assets/cutx.ico -arch amd64 -o windows_amd64.syso
	@echo "✓ Windows icon resource generated: windows_amd64.syso"

darwin-amd64: $(BUILDDIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/cutx-darwin-amd64 .

darwin-arm64: $(BUILDDIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/cutx-darwin-arm64 .

windows-amd64: $(BUILDDIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/cutx-windows-amd64.exe .

linux-amd64: $(BUILDDIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/cutx-linux-amd64 .

linux-arm64: $(BUILDDIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/cutx-linux-arm64 .

# Build for current platform
build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o cutx .

clean:
	rm -rf $(BUILDDIR) cutx windows_amd64.syso
