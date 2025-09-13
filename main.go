package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// 获取当前工作目录作为根目录
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前目录:", err)
	}
	
	log.Printf("文件服务器启动，根目录: %s", rootDir)
	log.Printf("服务器运行在: http://localhost:8080")
	log.Printf("使用 curl http://localhost:8080/your-file.sh 下载文件")
	
	// 创建文件服务器处理器
	fileServer := http.FileServer(http.Dir(rootDir))
	
	// 设置路由
	http.Handle("/", http.StripPrefix("/", fileServer))
	
	// 启动服务器
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}
	
	log.Fatal(http.ListenAndServe(port, nil))
}
