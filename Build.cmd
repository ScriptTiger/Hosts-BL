@echo off

set app=hosts-bl

if not exist Release md Release

set GOARCH=amd64

call :Build

set GOARCH=386

call :Build

pause

exit /b

:Build

set GOOS=windows
set file=%app%_%GOOS%_%GOARCH%.exe
call :Build_OS

set GOOS=linux
set file=%app%_%GOOS%_%GOARCH%
call :Build_OS

if %GOARCH% == 386 exit /b

set GOOS=darwin
set file=%app%_%GOOS%_%GOARCH%.app
call :Build_OS

exit /b

:Build_OS

echo Building Release/%file%...
go build -ldflags="-s -w" -o "Release/%file%"
exit /b