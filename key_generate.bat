@echo off
setlocal

:: 设置密钥文件名
set "PRIVATE_KEY=private.pem"
set "PUBLIC_KEY=public.pem"

:: 检查公钥是否存在
if exist "%PUBLIC_KEY%" (
    echo 公钥已存在：%PUBLIC_KEY%
    echo 跳过生成。
) else (
    echo 公钥不存在，正在通过 Git Bash 生成 RSA 密钥对...
    
    :: 调用 Git Bash 执行 OpenSSL 命令
    bash -c "openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out '%PRIVATE_KEY%' && openssl rsa -pubout -in '%PRIVATE_KEY%' -out '%PUBLIC_KEY%'"
    
    if errorlevel 1 (
        echo 错误：密钥生成失败。请确保已安装 Git for Windows 并将 Git Bash 加入系统 PATH。
        exit /b 1
    ) else (
        echo 密钥对生成成功！
        echo 私钥：%PRIVATE_KEY%
        echo 公钥：%PUBLIC_KEY%
    )
)

pause