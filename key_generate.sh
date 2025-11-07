#!/bin/bash

# 密钥文件名
PRIVATE_KEY="private.pem"
PUBLIC_KEY="public.pem"

# 检查公钥是否存在
if [ -f "$PUBLIC_KEY" ]; then
    echo "公钥已存在：$PUBLIC_KEY"
    echo "跳过生成。"
else
    echo "公钥不存在，正在生成 RSA 密钥对（2048 位）..."

    # 生成私钥
    if openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out "$PRIVATE_KEY"; then
        # 从私钥提取公钥
        if openssl rsa -pubout -in "$PRIVATE_KEY" -out "$PUBLIC_KEY"; then
            echo "✅ 密钥对生成成功！"
            echo "私钥：$PRIVATE_KEY"
            echo "公钥：$PUBLIC_KEY"
        else
            echo "❌ 提取公钥失败。"
            exit 1
        fi
    else
        echo "❌ 私钥生成失败。"
        exit 1
    fi
fi