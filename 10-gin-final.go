package main

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)
import sqlite "github.com/ncruces/go-sqlite3/gormlite"


// ========== JWT配置 ==========

var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// 生成Token
func generateToken(username string) (string, error) {
    claims := Claims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// 解析Token
func parseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, err
}

// ========== 认证中间件 ==========

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
            c.Abort()
            return
        }

        // 去掉 "Bearer " 前缀
        tokenString := authHeader
        if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
            tokenString = authHeader[7:]
        }

        claims, err := parseToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "无效token"})
            c.Abort()
            return
        }

        //将用户名存入上下文
        c.Set("username", claims.Username)
        c.Next()
    }
}

// ========== 数据库模型 ==========

var db *gorm.DB

type User struct {
    ID          uint    `json:"id" gorm:"primaryKey"`
    Username    string  `json:"username" gorm:"unique;not null"`
    Password    string  `json:"-" gorm:"not null"`  // json 忽略
}

type Article struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    Title       string      `json:"title" gorm:"not null"`
    Content     string      `json:"content"`
    Author      string      `json:"author"`
    CategoryID  *uint       `json:"category_id"`
    Category    Category    `json:"category" gorm:"foreignKey:CategoryID"`
    Comment     []Comment   `json:"comment" gorm:"foreignKey:ArticleID"`
}

// ========== 评论模型 ==========

type Comment struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    Content     string      `json:"content" gorm:"not null"`
    Author      string      `json:"author"`
    ArticleID   uint        `json:"article_id" gorm:"not null"`
    CreatedAt   time.Time   `json:"created_at"`
}

// ========== 分类模型 ==========
type Category struct {
    ID      uint    `json:"id" gorm:"primaryKey"`
    Name    string  `json:"name" gorm:"unique;not null"`
}

// ========== 密码处理 ==========

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// ========== Handler ==========

func register(c *gin.Context) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 检查用户是否存在
    var existing User
    if db.Where("username = ?", req.Username).First(&existing).RowsAffected > 0 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
    return
    }

    // 加密密码
    hashed, _ := hashPassword(req.Password)
    user := User{Username: req.Username, Password: hashed}
    db.Create(&user)

    c.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}

func login(c *gin.Context) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user User
    db.Where("username = ?", req.Username).First(&user)

    if user.ID == 0 || !checkPassword(req.Password, user.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        return
    }

    token, _ := generateToken(user.Username)
    c.JSON(http.StatusOK, gin.H{
        "access_token": token,
        "token_type": "bearer",
    })
}

func getMe(c *gin.Context) {
    username, _ := c.Get("username")
    c.JSON(http.StatusOK, gin.H{"user": username})
}

func createArticle(c *gin.Context) {
    var article Article
    if err := c.ShouldBindJSON(&article); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 从上下文获取当前用户
    username, _ := c.Get("username")
    article.Author = username.(string)

    db.Create(&article)
    c.JSON(http.StatusCreated, article)
}

// 更新文章
func updateArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if db.First(&article, id).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
    }

    var updateData Article
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 如果更新分类,验证分类存在
    if updateData.CategoryID != nil && *updateData.CategoryID != 0 {
        var category Category
        if db.First(&category, *updateData.CategoryID).Error != nil {
            c.JSON(404, gin.H{"error": "分类不存在"})
            return
        }
    }

    // 保留原作者, 不允许通过更新修改作者
    updateData.Author = article.Author

    db.Model(&article).Updates(updateData)
    c.JSON(200, article)
}

// 删除文章 (级联删除评论)
func deleteArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if db.First(&article, id).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    // 手动删除关联评论 (确保干净)
    db.Where("article_id = ?", id).Delete(&Comment{})

    db.Delete(&article)
    c.JSON(200, gin.H{"message": "删除成功"})
}

// 创建评论 (需要认证)
func createComment(c *gin.Context) {
    var comment Comment
    if err := c.ShouldBindJSON(&comment); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 验证文章存在
    var article Article
    if db.First(&article, comment.ArticleID).Error != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
        return
    }

    // 获取当前用户
    username, _ := c.Get("username")
    comment.Author = username.(string)

    db.Create(&comment)
    c.JSON(http.StatusCreated, comment)
}

// 获取文章评论 (公开)
func getArticleComments(c *gin.Context) {
    articleID := c.Param("id")

    var comments []Comment
    db.Where("article_id = ?", articleID).Find(&comments)

    c.JSON(http.StatusOK, comments)
}

// 创建分类
func createCategory(c *gin.Context) {
    var category Category
    if err := c.ShouldBindJSON(&category); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&category)
    c.JSON(http.StatusCreated, category)
}

// 更新分类
func updateCategory(c *gin.Context) {
    id := c.Param("id")
    var category Category
    if db.First(&category, id).Error != nil {
        c.JSON(404, gin.H{"error": "分类不存在"})
        return
    }

    var updateData Category
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    db.Model(&category).Updates(updateData)
    c.JSON(200, category)
}

// 删除分类
func deleteCategory(c *gin.Context) {
    id := c.Param("id")
	var category Category
	if db.First(&category, id).Error != nil {
		c.JSON(404, gin.H{"error": "分类不存在"})
		return
	}

	// 关联文章category_id设为NULL
	db.Model(&Article{}).Where("category_id = ?", id).Update("category_id", nil)

	db.Delete(&category)
	c.JSON(200, gin.H{"message": "删除成功"})
}

// 获取分类列表
func listCategories(c *gin.Context) {
    var categories []Category
    db.Find(&categories)
    c.JSON(http.StatusOK, categories)
}

// 获取分类下的文章
func getCategoryArticles(c *gin.Context) {
    categoryID := c.Param("id")

    var articles []Article
    db.Where("category_id = ?", categoryID).Find(&articles)

    c.JSON(http.StatusOK, articles)
}


// ========== 主函数 ==========

func main() {
    // 初始化数据库
    var err error
    db, err = gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&User{}, &Article{}, &Comment{}, &Category{})

    r := gin.Default()

    // 公开路由
    r.POST("/register", register)
    r.POST("/login", login)
    r.GET("/article/:id/comments", getArticleComments)
    r.GET("/categories", listCategories)
    r.GET("/categories/:id/articles", getCategoryArticles)

    //需要认证的路由
    auth := r.Group("/")
    auth.Use(JWTAuth())
    {
        auth.GET("/me", getMe)
        auth.POST("/articles", createArticle)
        auth.PUT("/articles/:id", updateArticle)
        auth.DELETE("/articles/:id", deleteArticle)
        auth.POST("/comments", createComment)
        auth.POST("/categories", createCategory)
        auth.PUT("/categories/:id", updateCategory)
        auth.DELETE("/categories/:id", deleteCategory)
    }

    r.Run(":8080")
}
