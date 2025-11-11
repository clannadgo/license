# License DLL Python SDK

Python SDK for calling the license.dll generated from Golang.

## 功能特性

- 生成机器指纹
- 验证许可证
- 获取许可证详细信息

## 环境要求

- Python 3.6 或更高版本
- Windows 操作系统（DLL 仅支持 Windows）

## 安装

1. 将 license.dll 复制到项目目录或系统路径
2. 安装依赖（如果有）：
```bash
pip install -r requirements.txt
```

## 使用示例

### 生成机器指纹

```python
from license_dll import LicenseUtils

license_utils = LicenseUtils()
fingerprint = license_utils.generate_fingerprint()
print(f"机器指纹: {fingerprint}")
```

### 验证许可证

```python
from license_dll import LicenseUtils

license_utils = LicenseUtils()
public_key_path = "public.pem"
license_content = "...";  # 许可证内容

result = license_utils.verify_license(public_key_path, license_content)
if result.success:
    print("许可证验证成功")
else:
    print(f"许可证验证失败: {result.message}")
```

### 获取许可证详细信息

```python
from license_dll import LicenseUtils

license_utils = LicenseUtils()
public_key_path = "public.pem"
license_content = "...";  # 许可证内容

license_data = license_utils.get_license_data(public_key_path, license_content)
if license_data:
    print(f"客户: {license_data.customer}")
    print(f"颁发者: {license_data.issuer}")
    print(f"过期时间: {datetime.fromtimestamp(license_data.expires_at)}")
```

## 运行示例

1. 确保 license.dll 在当前目录或系统路径中
2. 运行示例代码：
```bash
python example.py
```

## 注意事项

1. 确保 license.dll 在系统路径中，或者在初始化 LicenseUtils 时指定 DLL 路径：
```python
license_utils = LicenseUtils(dll_path="path/to/license.dll")
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

所有方法都可能抛出异常，建议在调用时进行适当的错误处理：

```python
try:
    fingerprint = license_utils.generate_fingerprint()
    # 使用指纹
except Exception as e:
    # 处理错误
    print(f"生成指纹失败: {e}")
```

## API 参考

### LicenseUtils 类

#### 构造函数
- `LicenseUtils(dll_path=None)`: 初始化 LicenseUtils，可选指定 DLL 路径

#### 方法
- `generate_fingerprint()`: 生成机器指纹
- `verify_license(public_key_path, license_content)`: 验证许可证
- `get_license_data(public_key_path, license_content)`: 获取许可证数据
- `read_license_file(license_path)`: 读取许可证文件
- `is_license_expired(public_key_path, license_content)`: 检查许可证是否过期

### LicenseVerificationResult 类

#### 属性
- `result_code`: 结果码
- `message`: 结果消息
- `success`: 是否成功

### LicenseData 类

#### 属性
- `customer`: 客户名称
- `issuer`: 颁发者
- `fingerprint`: 指纹
- `expires_at`: 过期时间（Unix 时间戳）
- `issued_at`: 颁发时间（Unix 时间戳）
- `raw_data`: 原始数据字典