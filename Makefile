# Makefile for building this Go program for multiple OS targets (single entry point)

# Base binary name
BINARY := faire

# Create folder function
define MAKE_BIN_DIR
	@mkdir -p bin
endef

# Build for Windows (AMD64)
windows-amd64:
	@echo "Building for Windows with GOOS=windows, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" \
		go build -ldflags="-H=windowsgui -extldflags=-mwindows" -o bin/$(BINARY)-windows-amd64.exe ./cmd/

# Build for Windows (ARM64)
windows-arm64:
	@echo "Building for Windows (ARM64) with GOOS=windows, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 CC="zig cc -target aarch64-windows" \
	    go build -ldflags="-H=windowsgui -extldflags=-mwindows" -o bin/$(BINARY)-windows-arm64.exe ./cmd/

# Build for Linux (AMD64)
linux-amd64:
	@echo "Building for Linux with GOOS=linux, GOARCH=amd64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC="zig cc -target x86_64-linux" \
		go build -o bin/$(BINARY)-linux-amd64 ./cmd/

# Build for Linux (ARM64)
linux-arm64:
	@echo "Building for Linux (ARM64) with GOOS=linux, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC="zig cc -target aarch64-linux" \
		go build -o bin/$(BINARY)-linux-arm64 ./cmd/

# Build for darwin (ARM64)
darwin-arm64:
	@echo "Building for darwin (ARM64) with GOOS=darwin, GOARCH=arm64..."
	$(call MAKE_BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=clang \
		go build -o bin/$(BINARY)-darwin-arm64 ./cmd/

# Build all targets
all: windows-amd64 windows-arm64 linux-amd64 linux-arm64 darwin-arm64

# Clean target to remove generated binaries and bin folder if needed
clean:
	@echo "Cleaning generated binaries and logs..."
	@rm -rf logs 2>/dev/null
	@rm -rf internal/app/logs 2>/dev/null
	@rm -rf bin 2>/dev/null

.PHONY: native windows-amd64 windows-arm64 linux-amd64 linux-arm64 darwin-arm64 all clean
