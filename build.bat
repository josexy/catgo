@echo off
setlocal EnableDelayedExpansion

set NAME=catgo
set BINDIR=bin
set MODULE=github.com/josexy/catgo
set PACKAGE=main.go

if not exist "%BINDIR%" mkdir "%BINDIR%"

:: 1. Version (git describe)
git describe --abbrev=0 --tags HEAD >nul 2>&1
if %errorlevel% equ 0 (
    for /f "delims=" %%i in ('git describe --abbrev^=0 --tags HEAD') do set VERSION=%%i
) else (
    set VERSION=unknown
)

:: 2. Git Commit
for /f "delims=" %%i in ('git rev-parse --short HEAD') do set GIT_COMMIT=%%i

:: 3. Go Version
for /f "delims=" %%i in ('go version') do set GO_VERSION=%%i

:: 4. Build Time
for /f "delims=" %%i in ('powershell -NoProfile -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'"') do set BUILD_TIME=%%i

echo [INFO] Build Config:
echo   Version:    %VERSION%
echo   Commit:     %GIT_COMMIT%
echo   Build Time: %BUILD_TIME%
echo   Go Version: %GO_VERSION% 

set LDFLAGS=-w -s
set LDFLAGS=%LDFLAGS% -X '%MODULE%/version.Version=%VERSION%'
set LDFLAGS=%LDFLAGS% -X '%MODULE%/version.GitCommit=%GIT_COMMIT%'
set LDFLAGS=%LDFLAGS% -X '%MODULE%/version.BuildTime=%BUILD_TIME%'
set LDFLAGS=%LDFLAGS% -X '%MODULE%/version.GoVersion=%GO_VERSION%'

set GOBUILD_BASE=go build -trimpath -ldflags "%LDFLAGS%"

if "%1"=="" goto build
if "%1"=="build" goto build
if "%1"=="all" goto all
if "%1"=="clean" goto clean

if "%1"=="linux-amd64" goto linux-amd64
if "%1"=="linux-arm64" goto linux-arm64
if "%1"=="linux-armv5" goto linux-armv5
if "%1"=="linux-armv6" goto linux-armv6
if "%1"=="linux-armv7" goto linux-armv7
if "%1"=="darwin-amd64" goto darwin-amd64
if "%1"=="darwin-arm64" goto darwin-arm64
if "%1"=="windows-amd64" goto windows-amd64
if "%1"=="windows-arm64" goto windows-arm64

echo [ERR] Unknown target: %1
exit /b 1

:build
echo [BUILD] Native build...
set CGO_ENABLED=0
%GOBUILD_BASE% -o "%BINDIR%\%NAME%.exe" %PACKAGE%
goto :eof

:all
echo [BUILD] Building ALL targets...
call :linux-amd64
call :linux-arm64
call :darwin-amd64
call :darwin-arm64
call :windows-amd64
call :windows-arm64
goto :eof

:linux-amd64
echo [BUILD] linux-amd64
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-linux-amd64" %PACKAGE%
goto :eof

:linux-arm64
echo [BUILD] linux-arm64
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-linux-arm64" %PACKAGE%
goto :eof

:linux-armv5
echo [BUILD] linux-armv5
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=5
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-linux-armv5" %PACKAGE%
set GOARM=
goto :eof

:linux-armv6
echo [BUILD] linux-armv6
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=6
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-linux-armv6" %PACKAGE%
set GOARM=
goto :eof

:linux-armv7
echo [BUILD] linux-armv7
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=arm
set GOARM=7
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-linux-armv7" %PACKAGE%
set GOARM=
goto :eof

:darwin-amd64
echo [BUILD] darwin-amd64
set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=amd64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-darwin-amd64" %PACKAGE%
goto :eof

:darwin-arm64
echo [BUILD] darwin-arm64
set CGO_ENABLED=0
set GOOS=darwin
set GOARCH=arm64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-darwin-arm64" %PACKAGE%
goto :eof

:windows-amd64
echo [BUILD] windows-amd64
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-windows-amd64.exe" %PACKAGE%
goto :eof

:windows-arm64
echo [BUILD] windows-arm64
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=arm64
%GOBUILD_BASE% -o "%BINDIR%\%NAME%-windows-arm64.exe" %PACKAGE%
goto :eof

:clean
echo [CLEAN] Removing binaries...
if exist "%BINDIR%" (
    del /q "%BINDIR%\%NAME%-*" 2>nul
    del /q "%BINDIR%\%NAME%.exe" 2>nul
)
echo [CLEAN] Done.
goto :eof
