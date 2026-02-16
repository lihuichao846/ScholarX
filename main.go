package main

import (
	"log"

	"paper-scraper/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// 静态文件
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// API 接口
	r.GET("/search", api.SearchPapers)
	r.GET("/daily-summary", api.GetDailySummary)

	log.Println("Server starting on http://localhost:8000")
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
