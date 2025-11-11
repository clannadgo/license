@echo off
echo Building license shared libraries for multiple platforms...

REM Create output directory
if not exist "output" mkdir output

REM Organize dependencies first
echo.
echo Organizing dependencies...
go mod tidy

REM Windows AMD64
echo.
echo Building for Windows AMD64...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_windows_amd64.dll license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Windows AMD64
) else (
    echo Success: license_windows_amd64.dll
)

REM Windows ARM64
echo.
echo Building for Windows ARM64...
set GOOS=windows
set GOARCH=arm64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_windows_arm64.dll license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Windows ARM64
) else (
    echo Success: license_windows_arm64.dll
)

REM Linux AMD64
echo.
echo Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_linux_amd64.so license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Linux AMD64
) else (
    echo Success: license_linux_amd64.so
)

REM Linux ARM64
echo.
echo Building for Linux ARM64...
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_linux_arm64.so license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Linux ARM64
) else (
    echo Success: license_linux_arm64.so
)

REM macOS AMD64
echo.
echo Building for macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_darwin_amd64.dylib license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for macOS AMD64
) else (
    echo Success: license_darwin_amd64.dylib
)

REM macOS ARM64
echo.
echo Building for macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=1
go build -buildmode=c-shared -o output/license_darwin_arm64.dylib license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for macOS ARM64
) else (
    echo Success: license_darwin_arm64.dylib
)

REM Create generic library files
echo.
echo Creating generic library files...

REM Copy current platform library as generic library file to output directory
copy output\license_windows_amd64.dll output\license.dll > nul 2>&1
if %errorlevel% equ 0 (
    echo Created: license.dll (Windows AMD64)
)

echo.
echo Build completed.
echo Output files in 'output' directory:
dir /b output\*

pause