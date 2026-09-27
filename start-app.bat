@echo off
cd /d "%~dp0"
go run .
if errorlevel 1 pause
