$env:PATH = "C:\Program Files\Go\bin;C:\Users\user\go\bin;" + $env:PATH
Set-Location $PSScriptRoot
& "C:\Users\user\go\bin\wails.exe" dev
