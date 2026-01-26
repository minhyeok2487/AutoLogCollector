@echo off
cd /d "%~dp0"
call credentials.cmd

rem Get today's date (YYYY-MM-DD)
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set today=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%

rem Create log folder
if not exist "logs\%today%" mkdir "logs\%today%"

rem Execute commands for each server
for /f "tokens=1,2 delims=," %%a in (servers.csv) do (
    echo [%%a] %%b Processing...
    (echo y & type commands.txt) | plink.exe %user%@%%a -pw %password% >"logs\%today%\%%b.log" 2>&1
    echo [%%a] %%b Done
)
echo All tasks completed
