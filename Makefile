.PHONY: all build build-darwin build-windows clean run deps

APP_NAME := breath
BUILD_DIR := build
CMD_PATH := ./cmd/breath

# 默认构建当前平台
all: build

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

# 构建 macOS 版本（仅在 macOS 上执行）
build-darwin:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)

# 构建 Windows 版本（需要安装 mingw-w64 交叉编译器）
# macOS 上安装: brew install mingw-w64
build-windows:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)

# 构建所有平台
build-all: build-darwin build-windows

# 运行应用
run:
	CGO_ENABLED=1 go run $(CMD_PATH)

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)

# 安装依赖
deps:
	go mod tidy
	go mod download
