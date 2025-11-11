#!/bin/bash

echo "Building license shared libraries for multiple platforms using CGO+zig..."

# 创建输出目录
mkdir -p output

# 首先整理依赖
echo ""
echo "Organizing dependencies..."
go mod tidy

# Windows AMD64
echo ""
echo "Building for Windows AMD64..."
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
# Use native Windows compiler for Windows platform
go build -buildmode=c-shared -o output/license_windows_amd64.dll license_dll.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Windows AMD64"
else
    echo "Success: license_windows_amd64.dll"
fi

# Windows ARM64
echo ""
echo "Skipping Windows ARM64 build (requires special cross-compilation setup)"
echo "Note: Windows ARM64 build skipped due to cross-compilation limitations"

# Linux AMD64
echo ""
echo "Building for Linux AMD64..."
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1
export CC="zig cc -target x86_64-linux-gnu"
export CXX="zig c++ -target x86_64-linux-gnu"
go build -buildmode=c-shared -o output/license_linux_amd64.so license_dll.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Linux AMD64"
else
    echo "Success: license_linux_amd64.so"
fi

# Linux ARM64
echo ""
echo "Building for Linux ARM64..."
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=1
export CC="zig cc -target aarch64-linux-gnu"
export CXX="zig c++ -target aarch64-linux-gnu"
go build -buildmode=c-shared -o output/license_linux_arm64.so license_dll.go
if [ $? -ne 0 ]; then
    echo "Failed to build for Linux ARM64"
else
    echo "Success: license_linux_arm64.so"
fi

# macOS AMD64
echo ""
echo "Skipping macOS AMD64 build (requires macOS SDK)"
echo "Note: macOS AMD64 build skipped due to missing system headers"

# macOS ARM64
echo ""
echo "Skipping macOS ARM64 build (requires macOS SDK)"
echo "Note: macOS ARM64 build skipped due to missing system headers"

# 创建通用库文件（符号链接或复制）
echo ""
echo "Creating generic library files..."

# 检测当前平台
CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH=$(uname -m)

# 转换架构名称
case $CURRENT_ARCH in
    x86_64)
        CURRENT_ARCH="amd64"
        ;;
    aarch64|arm64)
        CURRENT_ARCH="arm64"
        ;;
esac

# 复制当前平台的库为通用库文件到output目录
case $CURRENT_OS in
    linux)
        if [ -f "output/license_linux_${CURRENT_ARCH}.so" ]; then
            cp "output/license_linux_${CURRENT_ARCH}.so" "output/license.so"
            echo "Created: license.so (Linux ${CURRENT_ARCH})"
        fi
        ;;
    darwin)
        if [ -f "output/license_darwin_${CURRENT_ARCH}.dylib" ]; then
            cp "output/license_darwin_${CURRENT_ARCH}.dylib" "output/license.dylib"
            echo "Created: license.dylib (macOS ${CURRENT_ARCH})"
        fi
        ;;
esac

echo ""
echo "Build completed."
echo "Output files in 'output' directory:"
ls -la output/

echo ""
echo "To build for a specific platform only, use:"
echo "  ./build_platform.sh <os> <arch>"
echo "Example: ./build_platform.sh linux amd64"