$response = Invoke-WebRequest -Uri "http://localhost:8080/example-download.sh"
Write-Host "HTTP状态码: $($response.StatusCode)"
Write-Host "文件内容:"
Write-Host $response.Content