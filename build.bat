@echo off
setlocal


set "go_executable=go"
set "go_executable_name=repupperium.exe"


set "destination_dir=build"

if not exist %destination_dir% (
    echo Creating the "%destination_dir%" directory...
    mkdir %destination_dir%
)


%go_executable% build -ldflags="-s" -o %destination_dir%\%go_executable_name%


if %errorlevel% neq 0 (
    echo Build failed.
    exit /b
)



cd %destination_dir%
.\%go_executable_name%


pause

endlocal