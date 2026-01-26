@echo off
call credentials.cmd
for /f %%i in (iplist.txt) do (
    echo [%%i] Processing...
    (echo y & type commands.txt) | plink.exe %user%@%%i -pw %password% >%%i.log 2>&1
    echo [%%i] Done
)
echo All tasks completed
