@echo off
echo ========================================
echo 记忆力测试 - 第一轮对话
echo ========================================
echo 请求: 系统有什么告警？
echo.

powershell -Command "$body = '{\"Id\":\"memory-test\",\"Question\":\"系统有什么告警？\"}'; Invoke-WebRequest -Uri http://localhost:6872/api/chat -Method POST -Headers @{'Content-Type' = 'application/json'} -Body $body -UseBasicParsing | Select-Object -ExpandProperty Content"

echo.
echo ========================================
echo 等待 10 秒...
echo ========================================
timeout /t 10 /nobreak > nul

echo.
echo ========================================
echo 记忆力测试 - 第二轮对话
echo ========================================
echo 请求: 这些告警怎么处理？
echo 注意: 这次应该记住上一轮的内容
echo.

powershell -Command "$body = '{\"Id\":\"memory-test\",\"Question\":\"这些告警怎么处理？\"}'; Invoke-WebRequest -Uri http://localhost:6872/api/chat -Method POST -Headers @{'Content-Type' = 'application/json'} -Body $body -UseBasicParsing | Select-Object -ExpandProperty Content"

echo.
echo ========================================
echo 测试完成
echo ========================================
pause