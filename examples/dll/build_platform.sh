#!/bin/bash

# 检查参数
if [ $# -ne 2 ]; then
    echo "Usage: $0 <os> <arch>"
    echo "Supported OS: windows, linux, darwin"
    echo "Supported Arch: amd64, arm64"
    echo "Example: $0 linux amd64"
    exit 1
fi

OS=$1
ARCH=$2

# 验证参数
case $OS in
    windows|linux|darwin)
        ;;
    *)
        echo "Error: Unsupported OS '$OS'"
        echo "Supported OS: windows, linux, darwin"
        exit 1
        ;;
esac

case $ARCH in
    amd64|arm64)
        ;;
    *)
        echo "Error: Unsupported architecture '$ARCH'"
        echo "Supported architectures: amd64, arm64"
        exit 1
        ;;
esac

# 确定输出文件名和扩展名
case $OS in
    windows)
        EXT="dll"
        ;;
    linux)
        EXT="so"
        ;;
    darwin)
        EXT="dylib"
        ;;
esac

OUTPUT_FILE="license_${OS}_${ARCH}.${EXT}"

echo "Building license shared library for $OS/$ARCH..."

# 创建输出目录
mkdir -p output

# 首先整理依赖
echo "Organizing dependencies..."
go mod tidy

# 设置环境变量
export GOOS=$OS
export GOARCH=$ARCH
export CGO_ENABLED=1

# 构建
go build -buildmode=c-shared -o output/$OUTPUT_FILE license_dll.go

if [ $? -eq 0 ]; then
    echo "Success: $OUTPUT_FILE"
    
    # 如果是当前平台，也创建通用库文件
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
    
    if [ "$OS" = "$CURRENT_OS" ] && [ "$ARCH" = "$CURRENT_ARCH" ]; then
        case $OS in
            windows)
                cp output/$OUTPUT_FILE output/license.dll
                echo "Created: license.dll (Windows $ARCH)"
                ;;
            linux)
                cp output/$OUTPUT_FILE output/license.so
                echo "Created: license.so (Linux $ARCH)"
                ;;
            darwin)
                cp output/$OUTPUT_FILE output/license.dylib
                echo "Created: license.dylib (macOS $ARCH)"
                ;;
        esac
    fi
else
    echo "Failed to build for $OS/$ARCH"
    exit 1
fi