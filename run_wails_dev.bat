@echo off
set PATH=C:\Program Files\Go\bin;%USERPROFILE%\go\bin;%PATH%
cd /d "%~dp0"
wails dev
