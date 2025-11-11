# License DLL Java SDK

Java SDK for calling the license.dll generated from Golang.

## 功能特性

- 生成机器指纹
- 验证许可证
- 获取许可证详细信息

## 环境要求

- Java 11 或更高版本
- Maven 3.6 或更高版本
- Windows 操作系统（DLL 仅支持 Windows）

## 安装

1. 将 license.dll 复制到系统路径或项目目录
2. 添加依赖到 pom.xml：

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

## 使用示例

### 生成机器指纹

```java
import com.example.license.LicenseUtils;

String fingerprint = LicenseUtils.generateFingerprint();
System.out.println("机器指纹: " + fingerprint);
```

### 验证许可证

```java
import com.example.license.LicenseUtils;
import com.example.license.LicenseUtils.LicenseVerificationResult;

String publicKeyPath = "public.pem";
String licenseContent = "..."; // 许可证内容

LicenseVerificationResult result = LicenseUtils.verifyLicense(publicKeyPath, licenseContent);
if (result.isSuccess()) {
    System.out.println("许可证验证成功");
} else {
    System.out.println("许可证验证失败: " + result.getMessage());
}
```

### 获取许可证详细信息

```java
import com.example.license.LicenseUtils;
import com.example.license.LicenseUtils.LicenseData;

String publicKeyPath = "public.pem";
String licenseContent = "..."; // 许可证内容

LicenseData data = LicenseUtils.getLicenseData(publicKeyPath, licenseContent);
System.out.println("客户: " + data.getCustomer());
System.out.println("颁发者: " + data.getIssuer());
System.out.println("过期时间: " + new Date(data.getExpiresAt() * 1000));
```

## 运行示例

1. 编译项目：
```bash
mvn clean package
```

2. 运行示例：
```bash
java -cp target/license-dll-java-1.0.0.jar com.example.license.examples.LicenseDLLExample
```

或者使用 Maven 运行：
```bash
mvn exec:java -Dexec.mainClass="com.example.license.examples.LicenseDLLExample"
```

## 注意事项

1. 确保 license.dll 在系统路径中，或者通过系统属性 `jna.library.path` 指定路径：
```java
System.setProperty("jna.library.path", "path/to/dll");
```

2. 许可证验证需要公钥文件（public.pem）和许可证内容（license.lic）

3. 许可证验证结果码：
   - 0: 成功
   - 1: 无效的公钥
   - 2: 无效的许可证
   - 3: 许可证已过期
   - 4: 指纹不匹配
   - 5: 内部错误

## 错误处理

所有方法都可能抛出 RuntimeException，建议在调用时进行适当的错误处理：

```java
try {
    String fingerprint = LicenseUtils.generateFingerprint();
    // 使用指纹
} catch (RuntimeException e) {
    // 处理错误
    System.err.println("生成指纹失败: " + e.getMessage());
}
```