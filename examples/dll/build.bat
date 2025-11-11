@echo off
echo Building license DLL for Windows...

REM Create output directory
if not exist "output" mkdir output

REM Set environment variables
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1

REM Build DLL
go build -buildmode=c-shared -o output/license.dll license_dll.go

echo Build completed.
echo Output files in 'output' directory:
echo - license.dll (DLL file)
echo - license.h (Header file)

pause