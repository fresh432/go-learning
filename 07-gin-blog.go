package main

import(
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)
import sqlite "github.com/ncruces/go-sqlite3/gormlite"

// ========== 模型 ==========

type Article struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    Title       string      `json:"title" gorm:"not null"`
    Content     string      `json:"content"`
    Author      string      `json:"author" gorm:"default:'匿名'"`
    CategoryID  *uint       `json:"category_id"`
    CreatedAt   time.Time   `json:"created_at"`
}

type Category struct {
    ID      uint    `json:"id" gorm:"primaryKey"`
    Name    string  `json:"name" gorm:"unique;not null"`
}

// ========== 数据库 ==========

var db *gorm.DB

func initDB() {
    var err error
    db, err = gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&Article{}, &Category{})
}

// ========== Handler ==========

func listArticles(c *gin.Context) {
    var articles []Article
    db.Find(&articles)
    c.JSON(http.StatusOK, articles)
}

func getArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if result := db.First(&article, id); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
    }
    c.JSON(http.StatusOK, article)
}

func createArticle(c *gin.Context) {
    var article Article
    if err := c.ShouldBindJSON(&article); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&article)
    c.JSON(http.StatusCreated, article)
}

func updateArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if result := db.First(&article, id); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
        return
    }

    var updateData Article
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db.Model(&article).Updates(updateData)
    c.JSON(http.StatusOK, article)
}

func deleteArticle(c *gin.Context) {
    id := c.Param("id")
    if result := db.Delete(&Article{}, id); result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ========== 分类路由 ==========

func listCategories(c *gin.Context) {
    var categories []Category
    db.Find(&categories)
    c.JSON(http.StatusOK, categories)
}

func createCategory(c *gin.Context) {
    var category Category
    if err := c.ShouldBindJSON(&category); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&category)
    c.JSON(http.StatusCreated, category)
}

func getCategoryArticles(c *gin.Context) {
    id := c.Param("id")
    var articles []Article
    db.Where("category_id = ?", id).Find(&articles)
    c.JSON(http.StatusOK, articles)
}

// ========== 主函数 ==========

func main() {
    initDB()

    r := gin.Default()

    // 文章路由
    r.GET("/articles", listArticles)
    r.GET("/articles/:id", getArticle)
    r.POST("/articles", createArticle)
    r.PUT("/articles/:id", updateArticle)
    r.DELETE("/articles/:id", deleteArticle)

    // 分类路由
    r.GET("/categories", listCategories)
    r.POST("/categories", createCategory)
    r.GET("/categories/:id/articles", getCategoryArticles)

    r.Run(":8080")
}

