@echo off
chcp 65001 >nul
title DDNS-Updater 快速安装

echo ============================================
echo   DDNS-Updater 快速安装
echo   传奇服务端动态IP自动更新工具
echo ============================================
echo.

:: Check if exe exists
if not exist "%~dp0ddns-updater.exe" (
    echo ❌ 未找到 ddns-updater.exe
    echo 请先从 GitHub Releases 下载程序文件
    echo.
    pause
    exit /b 1
)

:: Generate config if not exists
if not exist "%~dp0config.yaml" (
    echo 正在生成配置文件...
    "%~dp0ddns-updater.exe" init
    echo.
) else (
    echo 配置文件已存在，跳过生成
    echo.
)

:: Open config for editing
echo 正在打开配置文件，请修改以下内容：
echo   1. file_updaters.path - 引擎配置文件路径
echo   2. file_updaters.old  - 当前公网IP
echo   3. db_updaters.path   - 微端网关数据库路径
echo   4. commands.args      - 重启命令
echo.
echo 修改完成后保存并关闭记事本
echo.
pause
notepad "%~dp0config.yaml"

:: Create startup shortcut
echo.
echo 是否创建开机自启快捷键？(Y/N)
set /p choice=
if /i "%choice%"=="Y" (
    echo 正在创建快捷方式...
    powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\DDNS-Updater.lnk'); $s.TargetPath = '%~dp0ddns-updater.exe'; $s.WorkingDirectory = '%~dp0'; $s.Save()"
    echo ✅ 开机自启快捷方式已创建
)

echo.
echo ============================================
echo 安装完成！
echo.
echo 启动方式：
echo   双击 ddns-updater.exe
echo   或运行: 快速启动.bat
echo ============================================
echo.

:: Create quick start bat
echo @echo off > "%~dp0快速启动.bat"
echo chcp 65001 ^>nul >> "%~dp0快速启动.bat"
echo title DDNS-Updater >> "%~dp0快速启动.bat"
echo echo DDNS-Updater 正在运行... >> "%~dp0快速启动.bat"
echo echo 按 Ctrl+C 停止 >> "%~dp0快速启动.bat"
echo echo. >> "%~dp0快速启动.bat"
echo "%~dp0ddns-updater.exe" >> "%~dp0快速启动.bat"

echo 已创建 快速启动.bat
echo.
pause
