@echo off
call credentials.cmd

for /f %%i in (iplist.txt) do (
    echo [%%i] Running commands...
    plink.exe %user%@%%i -pw %password% < commands.txt >%%i.log
    echo [%%i] Backing up config...
    plink.exe %user%@%%i -pw %password% "sh run" >%%i.config
    echo [%%i] Done
    echo.
)
echo All tasks completed
pause
