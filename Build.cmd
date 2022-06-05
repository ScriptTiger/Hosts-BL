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
set include=include_windows.go
call :Build_OS

set GOOS=linux
set file=%app%_%GOOS%_%GOARCH%
set include=include_other.go
call :Build_OS

if %GOARCH% == 386 exit /b

set GOOS=darwin
set file=%app%_%GOOS%_%GOARCH%.app
set include=include_other.go
call :Build_OS

exit /b

:Build_OS

echo Building Release/%file%...
go build -ldflags="-s -w" -o "Release/%file%" %app%.go %include%
set include=

exit /b