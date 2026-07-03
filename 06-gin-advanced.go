package main

import(
    "fmt"
    "time"
    "github.com/gin-gonic/gin"
    "net/http"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Logger中间件: 记录请求信息
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        // 执行后续handler
        c.Next()

        // 记录耗时
        fmt.Printf("[GIN] %s %s | 状态: %d | 耗时: %v\n",
        c.Request.Method,
        path,
        c.Writer.Status(),
        time.Since(start),
        )
    }
}

// 认证中间件: 简单版
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未授权"})
            c.Abort() // 终止后续处理
            return
        }
        c.Next()
    }
}

// func main() {
//     r := gin.Default()
//
//     // 全局使用日志中间件
//     r.Use(Logger())
//
//     // 公开路由
//     r.GET("/ping", func(c *gin.Context) {
//         c.JSON(200, gin.H{"message": "pong"})
//     })
//
//     // 需要认证的路由组
//     auth := r.Group("/api")
//     auth.Use(Auth())
//     {
//         auth.GET("/profile", func(c *gin.Context) {
//             c.JSON(200, gin.H{"user": "admin"})
//         })
//     }
//
//     r.Run(":8080")
// }

// func main() {
//     r := gin.Default()
//
//     // API v1 分组
//     v1 := r.Group("/api/v1")
//     {
//         // 文章路由
//         articles := v1.Group("/articles")
//         {
//             articles.GET("", listArticles)      // GET /api/v1/articles
//             articles.GET("/:id", getArticle)    // GET /api/v1/articles/1
//             articles.POST("", createArticle)    // POST /api/v1/articles
//         }
//
//         // 用户路由
//         users := v1.Group("/users")
//         {
//             users.POST("/register", register)
//             users.POST("/login", login)
//         }
//     }
//
//     // API v2 分组 (未来扩展)
//     v2 := r.Group("/api/v2")
//     {
//         v2.GET("/articles", listArticlesV2)
//     }
//
//     r.Run(":8080")
// }
//
// func listArticles(c *gin.Context) {
//     c.JSON(200, gin.H{"articles": []string{"articles1", "articles2"}})
// }
//
// func getArticle(c *gin.Context) {
//     id := c.Param("id")
//     c.JSON(200, gin.H{"id": id})
// }
//
// func createArticle(c *gin.Context) {
//     c.JSON(201, gin.H{"message": "创建成功"})
// }
//
// func register(c *gin.Context) {
//     c.JSON(200, gin.H{"message": "注册成功"})
// }
//
// func login(c *gin.Context) {
//     c.JSON(200, gin.H{"tokon": "xxx"})
// }
//
// func listArticlesV2(c *gin.Context) {
//     c.JSON(200, gin.H{"articles": "v2版本"})
// }

// ========== 模型定义 ==========
type Article struct {
    ID      uint    'json:"id" gorm:"primaryKey"'
    Title   string  'json:"title" gorm:"not null"'
    Content string  'json:"content"'
    Author  string  'json:"author" gorm:"default:'匿名'"'
}

// ========== 数据库初始化 ==========
var db *gorm.DB

func initDB() {
    var err error
    db, err = gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移
    db.AutoMigrate(&Article{})

    // 添加测试数据
    var count int64
    db.Model(&Article{}).Count(&count)
    if count == 0 {
        db.Create(&Article{Title: "第一篇", Content: "Hello GORM", Author: "fresh432"})
        db.Create(&Article{Title: "第二篇", Content: "Hello World", Author: "fresh432"})
    }
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
        return
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

func deleteArticle(c *gin.Context) {
    id := c.Param("id")
    if result := db.Delete(&Article{}, id); result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ========== 主函数 ==========
func main() {
    initDB()

    r := gin.Default()

    // 路由
    r.GET(".articles", listArticles)
    r.GET(".articles/:id", getArticle)
    r.POST(".articles", createArticle)
    r.DELETE(".articles/:id", deleteArticle)

    r.Run(":8080")
}
