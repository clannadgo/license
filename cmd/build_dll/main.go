package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// 支持的平台和架构组合
type Platform struct {
	OS        string
	Arch      string
	Ext       string // 共享库扩展名
	BuildMode string // 构建模式
}

func main() {
	// 定义支持的平台
	platforms := []Platform{
		{"windows", "amd64", "dll", "c-shared"},
		{"windows", "arm64", "dll", "c-shared"},
		{"linux", "amd64", "so", "c-shared"},
		{"linux", "arm64", "so", "c-shared"},
		{"darwin", "amd64", "dylib", "c-shared"},
		{"darwin", "arm64", "dylib", "c-shared"},
	}

	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 设置输出目录
	outputDir := filepath.Join(currentDir, "examples", "dll")

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
		return
	}

	// 为每个平台构建共享库
	for _, platform := range platforms {
		fmt.Printf("\n正在构建 %s/%s 平台的共享库...\n", platform.OS, platform.Arch)

		// 设置环境变量
		os.Setenv("GOOS", platform.OS)
		os.Setenv("GOARCH", platform.Arch)
		os.Setenv("CGO_ENABLED", "1")

		// 构建命令
		dllSourcePath := filepath.Join(outputDir, "license_dll.go")
		libName := fmt.Sprintf("license_%s_%s.%s", platform.OS, platform.Arch, platform.Ext)
		libOutputPath := filepath.Join(outputDir, libName)

		cmd := exec.Command("go", "build", "-buildmode="+platform.BuildMode, "-o", libOutputPath, dllSourcePath)
		cmd.Dir = outputDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// 执行构建
		if err := cmd.Run(); err != nil {
			fmt.Printf("构建 %s/%s 共享库失败: %v\n", platform.OS, platform.Arch, err)
			continue
		}

		fmt.Printf("✓ %s/%s 共享库构建成功!\n", platform.OS, platform.Arch)
		fmt.Printf("  输出文件: %s\n", libOutputPath)

		// 检查是否有头文件
		headerPath := libOutputPath + ".h"
		if _, err := os.Stat(headerPath); err == nil {
			fmt.Printf("  头文件: %s\n", headerPath)
		}

		// 检查是否有静态库
		staticLibPath := libOutputPath + ".a"
		if _, err := os.Stat(staticLibPath); err == nil {
			fmt.Printf("  静态库: %s\n", staticLibPath)
		}
	}

	// 创建一个通用的符号链接或复制文件，方便使用
	fmt.Println("\n创建通用库文件...")

	// 检测当前平台
	currentOS := strings.ToLower(os.Getenv("GOOS"))
	if currentOS == "" {
		// 如果GOOS未设置，尝试从运行时获取
		if runtime.GOOS != "" {
			currentOS = runtime.GOOS
		} else {
			currentOS = "windows" // 默认为Windows
		}
	}

	currentArch := strings.ToLower(os.Getenv("GOARCH"))
	if currentArch == "" {
		if runtime.GOARCH != "" {
			currentArch = runtime.GOARCH
		} else {
			currentArch = "amd64" // 默认为amd64
		}
	}

	// 查找匹配当前平台的库文件
	var currentPlatformLib string
	for _, platform := range platforms {
		if platform.OS == currentOS && platform.Arch == currentArch {
			currentPlatformLib = fmt.Sprintf("license_%s_%s.%s", platform.OS, platform.Arch, platform.Ext)
			break
		}
	}

	if currentPlatformLib != "" {
		// 确定通用库文件名
		var genericLibName string
		switch currentOS {
		case "windows":
			genericLibName = "license.dll"
		case "linux":
			genericLibName = "license.so"
		case "darwin":
			genericLibName = "license.dylib"
		default:
			genericLibName = "license." + platform.Ext
		}

		// 复制或创建符号链接
		sourcePath := filepath.Join(outputDir, currentPlatformLib)
		targetPath := filepath.Join(outputDir, genericLibName)

		// 尝试创建符号链接，如果失败则复制文件
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			// 符号链接创建失败，尝试复制文件
			if copyFile(sourcePath, targetPath) == nil {
				fmt.Printf("✓ 创建通用库文件: %s\n", targetPath)
			} else {
				fmt.Printf("✗ 创建通用库文件失败\n")
			}
		} else {
			fmt.Printf("✓ 创建通用库文件符号链接: %s\n", targetPath)
		}
	}

	fmt.Println("\n所有平台构建完成!")
}

// 复制文件
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
