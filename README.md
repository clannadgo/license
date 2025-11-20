# 许可证管理系统

这是一个完整的许可证管理系统，包括后端服务、前端界面和多平台SDK。

## 目录结构

- `cmd/` - 命令行工具
  - `gin/` - 后端服务主程序  -- 主体程序，后续用于管理许可证，请妥善保管您的私钥，公钥需要发放给客户
- `examples/` - 示例代码
  - `dll/` - 多平台共享库示例
  - `java/` - Java SDK示例
  - `python/` - Python SDK示例
- `frontend/` - 前端Vue.js应用-- npm install && npm run dev启动，或者进行打包部署
- `internal/` - 内部包
  - `config/` - 配置管理
  - `database/` - 数据库操作
  - `hwid/` - 硬件ID生成
  - `license/` - 许可证验证中间件

## 快速开始

### 1. 生成密钥对

首先，需要生成RSA密钥对用于签名和验证许可证：

```bash
key_generate.bat  生成公钥和私钥
```

### 2. 启动后端服务

```bash
go run cmd/gin/main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 3. 启动前端开发服务器

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器将在 `http://localhost:5174` 启动。

## 系统使用说明

### SDK使用

#### Go SDK

```go
参考 examples/go/example_usage.go
```

#### 多平台共享库

在 `examples/dll` 目录中提供了多平台共享库的构建脚本：

**Windows:**
```bash
cd examples/dll
build_all.bat
```

**Linux/macOS:**
```bash
cd examples/dll
chmod +x build_all.sh
./build_all.sh
```

构建完成后，`output` 目录将包含对应编译好的dll文件，该dll文件用于给非go语言的后端服务集成license使用

#### C/C++ SDK

```c
暂未调试，后续会调试并更新
```

#### Java SDK

```java
暂未调试，后续会调试并更新
```

#### Python SDK

```python
参考 examples/python/example_usage.py
```

## 配置说明

系统配置文件为 `config.json`：

```json
{
  "database": {
    "path": "license.db"
  },
  "server": {
    "port": 8080
  },
  "license": {
    "privateKeyPath": "private.pem",
    "publicKeyPath": "public.pem"
  }
}
```

## 部署说明

### 生产环境部署

1. 构建后端服务
   ```bash
   go build -o license.exe cmd/gin/main.go
   ```

2. 构建前端应用
   ```bash
   cd frontend
   npm run build
   ```

3. 使用反向代理（如Nginx）配置前端静态文件和API代理

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o license cmd/gin/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/license .
COPY --from=builder /app/config.json .
COPY --from=builder /app/private.pem .
COPY --from=builder /app/public.pem .
EXPOSE 8080
CMD ["./license"]
```

## 许可证格式

许可证采用JWS（JSON Web Signature）格式，包含以下信息：

```json
{
  "issuer": "许可证颁发者",
  "customer": "客户名称",
  "fingerprint": "机器指纹",
  "issuedAt": "颁发时间戳",
  "expiresAt": "过期时间戳"
}
```

## 安全注意事项

1. 私钥必须妥善保管，不可泄露
2. 许可证验证应在服务端进行，避免客户端绕过验证
3. 定期轮换密钥对，提高安全性
4. 机器指纹算法应考虑硬件变更的情况

## 故障排除

### 常见问题

1. **许可证验证失败**
   - 检查公钥是否正确
   - 检查许可证是否过期
   - 检查机器指纹是否匹配

2. **交叉编译问题**
   - 确保安装了对应平台的交叉编译工具
   - 对于非Windows平台，确保安装了zig编译器

3. **前端无法连接后端**
   - 检查后端服务是否启动
   - 检查CORS配置是否正确
   - 检查防火墙设置

## 贡献指南

欢迎提交Issue和Pull Request来改进这个项目。

## 许可证

本项目采用MIT许可证。