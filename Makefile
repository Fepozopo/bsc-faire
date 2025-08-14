# Makefile for building this Go program with CGO for multiple OS targets

# Base binary names
BINARY_CLI := faire-cli
BINARY_GUI := faire-gui

# Build CLI (default)
cli:
	$(call MAKE_BIN_DIR)
	go build -o bin/$(BINARY_CLI)-native ./cmd/main.go

# Build GUI (with build tag)
gui:
	$(call MAKE_BIN_DIR)
	go build -tags gui -o bin/$(BINARY_GUI)-native ./cmd/main_gui.go ./cmd/gui.go

# Create folder function
define MAKE_BIN_DIR
	@mkdir -p bin
endef

# Build for Windows (AMD64) CLI
cli-windows-amd64:
	@echo "Building CLI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -o bin/$(BINARY_CLI)-windows-amd64.exe ./cmd/main.go

# Build for Windows (AMD64) GUI
gui-windows-amd64:
	@echo "Building GUI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)-windows-amd64.exe ./cmd/main_gui.go ./cmd/gui.go

# Build for Windows (ARM64) CLI
cli-windows-arm64:
	@echo "Building CLI for Windows (ARM64) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-windows-arm64.exe ./cmd/main.go

# Build for Windows (ARM64) GUI
gui-windows-arm64:
	@echo "Building GUI for Windows (ARM64) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 CC="zig cc -target aarch64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)-windows-arm64.exe ./cmd/main_gui.go ./cmd/gui.go

# Build for Linux (AMD64) CLI
cli-linux-amd64:
	@echo "Building CLI for Linux with GOOS=linux, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -o bin/$(BINARY_CLI)-linux-amd64 ./cmd/main.go

# Build for Linux (AMD64) GUI
gui-linux-amd64:
	@echo "Building GUI for Linux with GOOS=linux, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" \
		go build -tags gui -o bin/$(BINARY_GUI)-linux-amd64 ./cmd/main_gui.go ./cmd/gui.go

# Build for Linux (ARM64) CLI
cli-linux-arm64:
	@echo "Building CLI for Linux (ARM64) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-linux-arm64 ./cmd/main.go

# Build for Linux (ARM64) GUI
gui-linux-arm64:
	@echo "Building GUI for Linux (ARM64) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" \
		go build -tags gui -o bin/$(BINARY_GUI)-linux-arm64 ./cmd/main_gui.go ./cmd/gui.go

# Build for darwin (ARM64) CLI
cli-darwin-arm64:
	@echo "Building CLI for darwin (ARM64) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-darwin-arm64 ./cmd/main.go

# Build for darwin (ARM64) GUI
gui-darwin-arm64:
	@echo "Building GUI for darwin (ARM64) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=clang \
		go build -tags gui -o bin/$(BINARY_GUI)-darwin-arm64 ./cmd/main_gui.go ./cmd/gui.go

# Combo targets to build both CLI and GUI for each platform/arch
windows-amd64: cli-windows-amd64 gui-windows-amd64
windows-arm64:  cli-windows-arm64  gui-windows-arm64
linux-amd64:   cli-linux-amd64   gui-linux-amd64
linux-arm64:    cli-linux-arm64    gui-linux-arm64
darwin-arm64:    cli-darwin-arm64    gui-darwin-arm64

# Build all CLI and GUI targets for all platforms
all: cli gui \
	cli-windows-amd64 gui-windows-amd64 \
	cli-windows-arm64 gui-windows-arm64 \
	cli-linux-amd64 gui-linux-amd64 \
	cli-linux-arm64 gui-linux-arm64 \
	cli-darwin-arm64 gui-darwin-arm64

# Clean target to remove generated binaries and bin folder if needed
clean:
	@echo "Cleaning generated binaries and logs..."
	@rm -rf logs 2>/dev/null
	@rm -rf internal/app/logs 2>/dev/null
	@rm -rf bin 2>/dev/null

.PHONY: cli gui \
	cli_windows_amd64 gui_windows_amd64 \
	cli-windows-arm64 gui-windows-arm64 \
	cli-linux-amd64 gui-linux-amd64 \
	cli-linux-arm64 gui-linux-arm64 \
	cli-darwin-arm64 gui-darwin-arm64 \
	cli-darwin-arm64 gui-darwin-arm64 \
	all clean
