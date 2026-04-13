.PHONY: all build build-darwin build-darwin-app build-windows clean run deps

APP_NAME := breath
BUILD_DIR := build
CMD_PATH := ./cmd/breath
BUNDLE_ID := com.breath.app

# 默认构建当前平台
all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

# 构建 macOS 二进制
build-darwin:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)

# 构建 macOS .app bundle（双击运行不会弹终端窗口）
build-darwin-app:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@mkdir -p $(BUILD_DIR)/Breath.app/Contents/MacOS
	@mkdir -p $(BUILD_DIR)/Breath.app/Contents/Resources
	@cp $(BUILD_DIR)/$(APP_NAME) $(BUILD_DIR)/Breath.app/Contents/MacOS/$(APP_NAME)
	@echo '<?xml version="1.0" encoding="UTF-8"?>\n\
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">\n\
<plist version="1.0">\n\
<dict>\n\
	<key>CFBundleExecutable</key>\n\
	<string>$(APP_NAME)</string>\n\
	<key>CFBundleIdentifier</key>\n\
	<string>$(BUNDLE_ID)</string>\n\
	<key>CFBundleName</key>\n\
	<string>Breath</string>\n\
	<key>CFBundleDisplayName</key>\n\
	<string>Breath</string>\n\
	<key>CFBundleVersion</key>\n\
	<string>1.0.0</string>\n\
	<key>CFBundleShortVersionString</key>\n\
	<string>1.0.0</string>\n\
	<key>CFBundlePackageType</key>\n\
	<string>APPL</string>\n\
	<key>LSMinimumSystemVersion</key>\n\
	<string>10.13</string>\n\
	<key>LSUIElement</key>\n\
	<true/>\n\
	<key>NSHighResolutionCapable</key>\n\
	<true/>\n\
</dict>\n\
</plist>' > $(BUILD_DIR)/Breath.app/Contents/Info.plist
	@echo "✅ 已生成 $(BUILD_DIR)/Breath.app"

# 构建 Windows 版本（纯 Go，不需要 CGO 和 GCC）
# -H windowsgui 隐藏命令行窗口
build-windows:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)

# 构建所有平台
build-all: build-darwin build-darwin-app build-windows

# 运行应用
run:
	go run $(CMD_PATH)

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)

# 安装依赖
deps:
	go mod tidy
	go mod download
