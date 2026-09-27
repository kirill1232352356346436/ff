@echo off
cd /d "%~dp0"
if not exist "server\.venv\Scripts\python.exe" (
    py -3 -m venv server\.venv
    if errorlevel 1 goto failed
)
server\.venv\Scripts\python.exe -m pip install -r server\requirements.txt
if errorlevel 1 goto failed
server\.venv\Scripts\python.exe -m uvicorn server.main:app --host 127.0.0.1 --port 8010
if errorlevel 1 goto failed
exit /b 0
:failed
pause
exit /b 1
