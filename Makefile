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

# Build for Windows (x86_64) CLI
windows-x86_64-cli:
	@echo "Building CLI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -o bin/$(BINARY_CLI)-windows-x86_64.exe ./cmd/main.go

# Build for Windows (x86_64) GUI
windows-x86_64-gui:
	@echo "Building GUI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)-windows-x86_64.exe ./cmd/main_gui.go ./cmd/gui.go

# Build for Windows (ARM64) CLI
windows-arm64-cli:
	@echo "Building CLI for Windows (ARM64) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-windows-arm64.exe ./cmd/main.go

# Build for Windows (ARM64) GUI
windows-arm64-gui:
	@echo "Building GUI for Windows (ARM64) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 CC="zig cc -target aarch64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)-windows-arm64.exe ./cmd/main_gui.go ./cmd/gui.go

# Build for Linux (x86_64) CLI
linux-x86_64-cli:
	@echo "Building CLI for Linux with GOOS=linux, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -o bin/$(BINARY_CLI)-linux-x86_64 ./cmd/main.go

# Build for Linux (x86_64) GUI
linux-x86_64-gui:
	@echo "Building GUI for Linux with GOOS=linux, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" \
		go build -tags gui -o bin/$(BINARY_GUI)-linux-x86_64 ./cmd/main_gui.go ./cmd/gui.go

# Build for Linux (ARM64) CLI
linux-arm64-cli:
	@echo "Building CLI for Linux (ARM64) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-linux-arm64 ./cmd/main.go

# Build for Linux (ARM64) GUI
linux-arm64-gui:
	@echo "Building GUI for Linux (ARM64) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" \
		go build -tags gui -o bin/$(BINARY_GUI)-linux-arm64 ./cmd/main_gui.go ./cmd/gui.go

# Build for macOS (ARM64) CLI
macos-arm64-cli:
	@echo "Building CLI for macOS (ARM64) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-macos-arm64 ./cmd/main.go

# Build for macOS (ARM64) GUI
macos-arm64-gui:
	@echo "Building GUI for macOS (ARM64) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=clang \
		go build -tags gui -o bin/$(BINARY_GUI)-macos-arm64 ./cmd/main_gui.go ./cmd/gui.go

# Combo targets to build both CLI and GUI for each platform/arch
windows-x86_64: windows-x86_64-cli windows-x86_64-gui
windows-arm64:  windows-arm64-cli  windows-arm64-gui
linux-x86_64:   linux-x86_64-cli   linux-x86_64-gui
linux-arm64:    linux-arm64-cli    linux-arm64-gui
macos-arm64:    macos-arm64-cli    macos-arm64-gui

# Build all CLI and GUI targets for all platforms
all: cli gui \
	windows-x86_64-cli windows-x86_64-gui \
	windows-arm64-cli windows-arm64-gui \
	linux-x86_64-cli linux-x86_64-gui \
	linux-arm64-cli linux-arm64-gui \
	macos-arm64-cli macos-arm64-gui

# Clean target to remove generated binaries and bin folder if needed
clean:
	@echo "Cleaning generated binaries and logs..."
	@rm -rf logs 2>/dev/null
	@rm -rf internal/app/logs 2>/dev/null
	@rm -rf bin 2>/dev/null

.PHONY: cli gui \
	windows-x86_64-cli windows-x86_64-gui \
	windows-arm64-cli windows-arm64-gui \
	linux-x86_64-cli linux-x86_64-gui \
	linux-arm64-cli linux-arm64-gui \
	macos-arm64-cli macos-arm64-gui \
	all clean
