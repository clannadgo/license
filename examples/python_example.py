import base64
import binascii
import json
import time
from pathlib import Path
import jwt
from cryptography.hazmat.primitives import serialization

# ------------------ 公钥加载 ------------------
def load_public_key(pem_path: str):
    with open(pem_path, "rb") as f:
        pem_data = f.read()
    pub = serialization.load_pem_public_key(pem_data)
    return pub

# ------------------ 验证 license ------------------
def verify_license_file(license_path: str, pub_key_path: str):
    pub = load_public_key(pub_key_path)

    # 读取 license
    with open(license_path, "r") as f:
        license_str = f.read().strip()

    # 验签 JWS
    try:
        payload = jwt.decode(
            license_str,
            key=pub,
            algorithms=["RS256"],
            options={"verify_exp": False}  # 我们自己处理 exp
        )
    except jwt.exceptions.InvalidSignatureError:
        raise ValueError("invalid license signature")

    # 校验过期时间（精确到分钟）
    exp = payload.get("exp")
    if exp:
        now_minute = int(time.time() / 60)
        exp_minute = int(exp / 60)
        if now_minute > exp_minute:
            raise ValueError("license expired")

    return payload

# ------------------ 示例 ------------------
if __name__ == "__main__":
    license_file = "license.lic"
    pem_file = "public.pem"

    payload = verify_license_file(license_file, pem_file)
    print("License valid")
    print("Customer:", payload.get("customer"))
    print("Expires:", payload.get("exp"))
    print("Fingerprint (hex):", payload.get("fingerprint"))
