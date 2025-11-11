# License DLL SDK 文档

本文档介绍如何使用从 Golang 生成的 license 共享库，以及如何在 Java 和 Python 中调用该库。

## 目录

1. [概述](#概述)
2. [生成共享库](#生成共享库)
3. [共享库 API 参考](#共享库-api-参考)
4. [Java SDK](#java-sdk)
5. [Python SDK](#python-sdk)
6. [示例代码](#示例代码)
7. [常见问题](#常见问题)

## 概述

License 共享库是一个从 Golang 代码生成的动态链接库，提供以下功能：

- 生成机器指纹
- 验证许可证
- 获取许可证详细信息

该共享库支持跨平台编译和使用，可在 Windows、Linux 和 macOS 上被 Java 和 Python 调用。

## 生成共享库

### 跨平台生成

1. 进入项目目录：
```bash
cd c:\work_place\code\license\license
```

2. 使用内置的生成脚本：
```bash
# Windows
go run cmd\build_dll\main.go

# 或使用跨平台构建脚本
cd examples\dll
build_all.bat      # Windows
build_all.sh       # Linux/macOS
```

### 手动生成

#### Windows

1. 设置环境变量：
```bash
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
```

2. 编译 DLL：
```bash
go build -buildmode=c-shared -o license.dll license_dll.go
```

#### Linux

1. 设置环境变量：
```bash
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1
```

2. 编译共享库：
```bash
go build -buildmode=c-shared -o license.so license_dll.go
```

#### macOS

1. 设置环境变量：
```bash
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1
```

2. 编译共享库：
```bash
go build -buildmode=c-shared -o license.dylib license_dll.go
```

#### ARM 架构支持

对于 ARM 架构（如 ARM64），只需将 GOARCH 设置为对应的架构：

```bash
# Linux ARM64
export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=1
go build -buildmode=c-shared -o license.so license_dll.go

# macOS Apple Silicon (ARM64)
export GOOS=darwin
export GOARCH=arm64
export CGO_ENABLED=1
go build -buildmode=c-shared -o license.dylib license_dll.go
```

## 共享库 API 参考

共享库提供以下导出函数：

### GenerateFingerprint

```c
char* GenerateFingerprint();
```

**功能**: 生成机器指纹

**返回值**: 机器指纹字符串（UTF-8 编码），需要调用 FreeString 释放内存

### VerifyLicense

```c
int VerifyLicense(const char* publicKeyPath, const char* licenseContent);
```

**功能**: 验证许可证

**参数**:
- `publicKeyPath`: 公钥文件路径（UTF-8 编码）
- `licenseContent`: 许可证内容（UTF-8 编码）

**返回值**:
- 0: 成功
- 1: 无效的公钥
- 2: 无效的许可证
- 3: 许可证已过期
- 4: 指纹不匹配
- 5: 内部错误

### GetLicenseData

```c
char* GetLicenseData(const char* publicKeyPath, const char* licenseContent);
```

**功能**: 获取许可证数据（JSON 格式）

**参数**:
- `publicKeyPath`: 公钥文件路径（UTF-8 编码）
- `licenseContent`: 许可证内容（UTF-8 编码）

**返回值**: 许可证数据的 JSON 字符串（UTF-8 编码），需要调用 FreeString 释放内存

### FreeString

```c
void FreeString(char* str);
```

**功能**: 释放由共享库分配的字符串内存

**参数**:
- `str`: 要释放的字符串指针

## Java SDK

### 环境要求

- Java 11 或更高版本
- Maven 3.6 或更高版本
- Windows、Linux 或 macOS 操作系统

### 安装

1. 添加依赖到 pom.xml：

```xml
<dependency>
    <groupId>net.java.dev.jna</groupId>
    <artifactId>jna</artifactId>
    <version>5.12.0</version>
</dependency>
<dependency>
    <groupId>com.fasterxml.jackson.core</groupId>
    <artifactId>jackson-databind</artifactId>
    <version>2.14.2</version>
</dependency>
```

2. 将对应平台的共享库复制到系统路径或项目目录：
   - Windows: license.dll
   - Linux: license.so
   - macOS: license.dylib

### 使用示例

```java
import com.example.license.LicenseUtils;

// 初始化
LicenseUtils licenseUtils = new LicenseUtils();

// 获取平台信息
String platformInfo = licenseUtils.getPlatformInfo();
System.out.println("平台信息: " + platformInfo);

// 生成机器指纹
String fingerprint = licenseUtils.generateFingerprint();
System.out.println("机器指纹: " + fingerprint);

// 验证许可证
String publicKeyPath = "public.pem";
String licenseContent = "..."; // 许可证内容

LicenseVerificationResult result = licenseUtils.verifyLicense(publicKeyPath, licenseContent);
if (result.isSuccess()) {
    System.out.println("许可证验证成功");
} else {
    System.out.println("许可证验证失败: " + result.getMessage());
}

// 获取许可证详细信息
LicenseData data = licenseUtils.getLicenseData(publicKeyPath, licenseContent);
System.out.println("客户: " + data.getCustomer());
System.out.println("颁发者: " + data.getIssuer());
System.out.println("过期时间: " + new Date(data.getExpiresAt() * 1000));
```

### 文件结构

```
examples/java/license_dll/
├── pom.xml
├── README.md
└── src/main/java/com/example/license/
    ├── LicenseDLL.java          # JNA 接口定义
    ├── LicenseUtils.java        # 工具类
    └── examples/
        └── LicenseDLLExample.java  # 示例代码
```

## Python SDK

### 环境要求

- Python 3.6 或更高版本
- Windows、Linux 或 macOS 操作系统

### 安装

1. 将对应平台的共享库复制到项目目录或系统路径：
   - Windows: license.dll
   - Linux: license.so
   - macOS: license.dylib
2. 安装依赖（如果有）：
```bash
pip install -r requirements.txt
```

### 使用示例

```python
from license_dll import LicenseUtils

# 初始化
license_utils = LicenseUtils()

# 获取平台信息
platform_info = license_utils.get_platform_info()
print(f"平台信息: {platform_info}")

# 生成机器指纹
fingerprint = license_utils.generate_fingerprint()
print(f"机器指纹: {fingerprint}")

# 验证许可证
public_key_path = "public.pem"
license_content = "...";  # 许可证内容

result = license_utils.verify_license(public_key_path, license_content)
if result.success:
    print("许可证验证成功")
else:
    print(f"许可证验证失败: {result.message}")

# 获取许可证详细信息
license_data = license_utils.get_license_data(public_key_path, license_content)
if license_data:
    print(f"客户: {license_data.customer}")
    print(f"颁发者: {license_data.issuer}")
    print(f"过期时间: {datetime.fromtimestamp(license_data.expires_at)}")
```

### 文件结构

```
examples/python/license_dll/
├── __init__.py
├── license_dll.py           # 主要模块
├── example.py              # 示例代码
├── requirements.txt        # 依赖列表
└── README.md              # 文档
```

## 示例代码

### Java 示例

完整示例代码位于 `examples/java/license_dll/src/main/java/com/example/license/examples/LicenseDLLExample.java`

运行示例：
```bash
cd examples/java/license_dll
mvn clean package
java -cp target/license-dll-java-1.0.0.jar com.example.license.examples.LicenseDLLExample
```

### Python 示例

完整示例代码位于 `examples/python/license_dll/example.py`

运行示例：
```bash
cd examples/python/license_dll
python example.py
```

## 常见问题

### 1. 共享库加载失败

**问题**: 无法加载共享库

**解决方案**:
- 确保共享库在系统路径中
- 或者在代码中指定共享库路径：
  - Java: `System.setProperty("jna.library.path", "path/to/library");`
  - Python: `LicenseUtils(library_path="path/to/license")`
- 确保使用与当前平台匹配的共享库：
  - Windows: license.dll
  - Linux: license.so
  - macOS: license.dylib

### 2. 许可证验证失败

**问题**: 许可证验证返回错误码

**解决方案**:
- 检查公钥文件是否存在且可读
- 检查许可证内容格式是否正确
- 确认机器指纹是否匹配
- 检查许可证是否已过期

### 3. 内存泄漏

**问题**: 长时间运行后内存使用增加

**解决方案**:
- 确保调用 FreeString 释放由共享库分配的内存
- Java 和 Python SDK 已自动处理内存释放

### 4. 编码问题

**问题**: 字符串显示乱码

**解决方案**:
- 确保所有字符串使用 UTF-8 编码
- 检查文件读取时是否指定了正确的编码

### 5. 跨平台兼容性

**问题**: 在不同平台上使用时遇到问题

**解决方案**:
- 确保为每个目标平台编译了对应的共享库
- 使用 `getPlatformInfo()` 方法检查当前平台信息
- 在 Linux/macOS 上可能需要设置 LD_LIBRARY_PATH 或 DYLD_LIBRARY_PATH 环境变量

### 6. ARM 架构支持

**问题**: 在 ARM 架构设备上无法使用

**解决方案**:
- 使用对应的 GOARCH 设置重新编译共享库：
  ```bash
  # Linux ARM64
  export GOOS=linux
  export GOARCH=arm64
  export CGO_ENABLED=1
  go build -buildmode=c-shared -o license.so license_dll.go
  
  # macOS Apple Silicon (ARM64)
  export GOOS=darwin
  export GOARCH=arm64
  export CGO_ENABLED=1
  go build -buildmode=c-shared -o license.dylib license_dll.go
  ```

### 7. 机器指纹不一致

**问题**: 同一台设备上不同平台生成的指纹不一致

**解决方案**:
- 这是正常现象，因为不同平台使用不同的方法获取机器ID
- 许可证验证时会自动使用当前平台的指纹生成方法
- 如需跨平台使用同一许可证，请在生成许可证时包含所有平台的指纹

## 更新日志

### v1.1.0 (2023-11-15)

- 添加跨平台支持（Windows、Linux、macOS）
- 添加 ARM 架构支持
- 更新 Java 和 Python SDK 以支持跨平台
- 添加平台信息获取功能

### v1.0.0 (2023-11-11)

- 初始版本发布
- 支持生成机器指纹
- 支持许可证验证
- 提供 Java 和 Python SDK
- 提供完整的示例代码和文档