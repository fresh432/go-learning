/*
Go Gin框架入门
学习时间：2026-07-02 20:35-21:05
来源：Gin官方文档
*/

package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// 文章结构体
type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var articles = []Article{
	{ID: 1, Title: "第一篇", Content: "Hello Gin"},
	{ID: 2, Title: "第二篇", Content: "Go Web框架"},
}

func main() {
	r := gin.Default()

	// GET /ping 测试
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// GET /articles 获取所有文章
	r.GET("/articles", func(c *gin.Context) {
		c.JSON(http.StatusOK, articles)
	})

	// GET /articles/:id 获取单篇文章
	r.GET("/articles/:id", func(c *gin.Context) {
		id := c.Param("id")
		// 简化处理，实际应转换int并查找
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	// POST /articles 创建文章
	r.POST("/articles", func(c *gin.Context) {
		var newArticle Article
		if err := c.ShouldBindJSON(&newArticle); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		articles = append(articles, newArticle)
		c.JSON(http.StatusCreated, newArticle)
	})

	// 运行
	r.Run(":8080")  // 默认8080
}