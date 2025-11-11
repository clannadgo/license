@echo off
echo Building license.dll...

REM 设置环境变量
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

REM 构建DLL
go build -buildmode=c-shared -o license.dll license_dll.go

echo Build completed.
echo Output files:
echo - license.dll (DLL文件)
echo - license.dll.h (头文件)
echo - license.dll.a (静态库)

pause