# Makefile for building this Go program with CGO for multiple OS targets

# Base binary names
BINARY_CLI := faire-cli
BINARY_GUI := faire-gui

# Build CLI (default)
cli:
	$(call MAKE_BIN_DIR)
	go build -o bin/$(BINARY_CLI) ./cmd/main.go

# Build GUI (with build tag)
gui:
	$(call MAKE_BIN_DIR)
	go build -tags gui -o bin/$(BINARY_GUI) ./cmd/main_gui.go ./cmd/gui.go


# Create folder function
define MAKE_BIN_DIR
	@mkdir -p bin
endef



# Build for Windows (x86_64) CLI
windows-x86_64-cli:
	@echo "Building CLI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
		go build -o bin/$(BINARY_CLI)_windows-x86_64.exe ./cmd/main.go

# Build for Windows (x86_64) GUI
windows-x86_64-gui:
	@echo "Building GUI for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)_windows-x86_64.exe ./cmd/main_gui.go ./cmd/gui.go


# Build for Windows (ARM) CLI
windows-arm-cli:
	@echo "Building CLI for Windows (ARM) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)_windows-arm.exe ./cmd/main.go

# Build for Windows (ARM) GUI
windows-arm-gui:
	@echo "Building GUI for Windows (ARM) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 CC="zig cc -target aarch64-windows" \
		go build -tags gui -o bin/$(BINARY_GUI)_windows-arm.exe ./cmd/main_gui.go ./cmd/gui.go


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


# Build for Linux (ARM) CLI
linux-arm-cli:
	@echo "Building CLI for Linux (ARM) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-linux-arm ./cmd/main.go

# Build for Linux (ARM) GUI
linux-arm-gui:
	@echo "Building GUI for Linux (ARM) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" \
		go build -tags gui -o bin/$(BINARY_GUI)-linux-arm ./cmd/main_gui.go ./cmd/gui.go


# Build for macOS (ARM) CLI
macos-arm-cli:
	@echo "Building CLI for macOS (ARM) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 \
		go build -o bin/$(BINARY_CLI)-macos-arm ./cmd/main.go

# Build for macOS (ARM) GUI
macos-arm-gui:
	@echo "Building GUI for macOS (ARM) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=clang \
		go build -tags gui -o bin/$(BINARY_GUI)-macos-arm ./cmd/main_gui.go ./cmd/gui.go


# Build all CLI and GUI targets for all platforms
all: cli gui \
	windows-x86_64-cli windows-x86_64-gui \
	windows-arm-cli windows-arm-gui \
	linux-x86_64-cli linux-x86_64-gui \
	linux-arm-cli linux-arm-gui \
	macos-arm-cli macos-arm-gui

# Clean target to remove generated binaries and bin folder if needed
clean:
	@echo "Cleaning generated binaries..."
	@rm -rf bin

.PHONY: cli gui \
	windows-x86_64-cli windows-x86_64-gui \
	windows-arm-cli windows-arm-gui \
	linux-x86_64-cli linux-x86_64-gui \
	linux-arm-cli linux-arm-gui \
	macos-arm-cli macos-arm-gui \
	all clean
