@echo off
echo Building license shared libraries for multiple platforms using CGO+zig...

REM Save current directory and change to script directory
pushd "%~dp0"

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
REM Use native Windows compiler for Windows platform
go build -buildmode=c-shared -o output/license_windows_amd64.dll .\license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Windows AMD64
) else (
    echo Success: license_windows_amd64.dll
)

REM Windows ARM64
echo.
echo Skipping Windows ARM64 build (requires special cross-compilation setup)
echo Note: Windows ARM64 build skipped due to cross-compilation limitations

REM Linux AMD64
echo.
echo Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=1
set CC=zig cc -target x86_64-linux-gnu
set CXX=zig c++ -target x86_64-linux-gnu
go build -buildmode=c-shared -o output/license_linux_amd64.so .\license_dll.go
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
set CC=zig cc -target aarch64-linux-gnu
set CXX=zig c++ -target aarch64-linux-gnu
go build -buildmode=c-shared -o output/license_linux_arm64.so .\license_dll.go
if %errorlevel% neq 0 (
    echo Failed to build for Linux ARM64
) else (
    echo Success: license_linux_arm64.so
)

REM macOS AMD64
echo.
echo Skipping macOS AMD64 build (requires macOS SDK)
echo Note: macOS AMD64 build skipped due to missing system headers

REM macOS ARM64
echo.
echo Skipping macOS ARM64 build (requires macOS SDK)
echo Note: macOS ARM64 build skipped due to missing system headers

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

REM Restore original directory
popd

pause