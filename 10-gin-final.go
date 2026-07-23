package main

import (
    "fmt"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"

    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

    _ "gin-demo/docs"  // 生成的docs包
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
    Avatar      string  `json:"avatar"`
    Bio         string  `json:"bio"`
}

type Article struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    Title       string      `json:"title" gorm:"not null"`
    Content     string      `json:"content"`
    Author      string      `json:"author"`
    CategoryID  *uint       `json:"category_id"`
    Category    Category    `json:"category" gorm:"foreignKey:CategoryID"`
    Comment     []Comment   `json:"comment" gorm:"foreignKey:ArticleID"`
    Likes       []Like      `json:"likes" gorm:"foreignKey:ArticleID"`
}

type Comment struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    Content     string      `json:"content" gorm:"not null"`
    Author      string      `json:"author"`
    ArticleID   uint        `json:"article_id" gorm:"not null"`
    CreatedAt   time.Time   `json:"created_at"`
}

type Category struct {
    ID      uint    `json:"id" gorm:"primaryKey"`
    Name    string  `json:"name" gorm:"unique;not null"`
}

// Like模型
type Like struct {
    ID          uint        `json:"id" gorm:"primaryKey"`
    UserID      uint        `json:"user_id" gorm:"not null"`
    ArticleID   uint        `json:"article_id" gorm:"not null"`
    CreatedAt   time.Time   `json:"created_at"`
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

// @Summary 用户注册
// @Description 创建一个新用户账号
// @Tags 用户
// @Accept json
// @Produce json
// @Param user body User true "用户信息"
// @Success 201 {object} map[string]string "注册成功"
// @Failure 400 {object} map[string]string "请求参数错误或用户名已存在"
// @Router /register [post]
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

// @Summary 用户登录
// @Description 使用用户名和密码登录，获取 JWT Token
// @Tags 用户
// @Accept json
// @Produce json
// @Param user body User true "登录凭证"
// @Success 200 {object} map[string]string "登录成功，返回Token"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "用户名或密码错误"
// @Router /login [post]
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

// @Summary 获取当前登录用户信息
// @Description 需要 Bearer Token 认证
// @Tags 用户
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "返回当前用户名"
// @Failure 401 {object} map[string]string "未授权"
// @Router /me [get]
func getMe(c *gin.Context) {
    username, _ := c.Get("username")
    c.JSON(http.StatusOK, gin.H{"user": username})
}

// @Summary 更新用户资料
// @Description 更新当前用户的头像和简介
// @Tags 用户
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} User
// @Router /users/me [put]
func updateProfile(c *gin.Context) {
    username, _ := c.Get("username")
    var user User
    if db.Where("username = ?", username).First(&user).Error != nil {
        c.JSON(404, gin.H{"error": "用户不存在"})
        return
    }

    var updateData struct {
        Avatar  string  `json:"avatar"`
        Bio     string  `json:"bio"`
    }

    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    db.Model(&user).Updates(updateData)
    c.JSON(200, user)
}

// @Summary 创建文章
// @Description 发布一篇新文章，需要认证
// @Tags 文章
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param article body Article true "文章详情"
// @Success 201 {object} Article "创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Router /articles [post]
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

// @Summary 更新文章
// @Description 根据ID更新文章信息，需要认证
// @Tags 文章
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Param article body Article true "更新后的文章信息"
// @Success 200 {object} Article "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "文章或分类不存在"
// @Router /articles/{id} [put]
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

// @Summary 删除文章
// @Description 根据ID删除文章，会级联删除评论，需要认证
// @Tags 文章
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "文章不存在"
// @Router /articles/{id} [delete]
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

// @Summary 创建评论
// @Description 为指定文章添加评论，需要认证
// @Tags 评论
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param comment body Comment true "评论内容"
// @Success 201 {object} Comment "评论成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "文章不存在"
// @Router /comments [post]
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

// @Summary 获取文章评论列表
// @Description 获取指定文章下的所有评论（公开接口）
// @Tags 评论
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {array} Comment "评论列表"
// @Router /article/{id}/comments [get]
func getArticleComments(c *gin.Context) {
    articleID := c.Param("id")

    var comments []Comment
    db.Where("article_id = ?", articleID).Find(&comments)

    c.JSON(http.StatusOK, comments)
}

// @Summary 创建分类
// @Description 添加一个新的文章分类，需要认证
// @Tags 分类
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param category body Category true "分类信息"
// @Success 201 {object} Category "创建成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Router /categories [post]
func createCategory(c *gin.Context) {
    var category Category
    if err := c.ShouldBindJSON(&category); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    db.Create(&category)
    c.JSON(http.StatusCreated, category)
}

// @Summary 更新分类
// @Description 修改分类名称，需要认证
// @Tags 分类
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "分类ID"
// @Param category body Category true "新的分类信息"
// @Success 200 {object} Category "更新成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "分类不存在"
// @Router /categories/{id} [put]
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

// @Summary 删除分类
// @Description 删除分类，并将关联文章的分类ID置空，需要认证
// @Tags 分类
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "分类ID"
// @Success 200 {object} map[string]string "删除成功"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 404 {object} map[string]string "分类不存在"
// @Router /categories/{id} [delete]
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

// @Summary 获取分类列表
// @Description 获取所有文章分类（公开接口）
// @Tags 分类
// @Produce json
// @Success 200 {array} Category "分类列表"
// @Router /categories [get]
func listCategories(c *gin.Context) {
    var categories []Category
    db.Find(&categories)
    c.JSON(http.StatusOK, categories)
}

// @Summary 获取分类下的文章
// @Description 获取指定分类下的所有文章（公开接口）
// @Tags 分类
// @Produce json
// @Param id path int true "分类ID"
// @Success 200 {array} Article "文章列表"
// @Router /categories/{id}/articles [get]
func getCategoryArticles(c *gin.Context) {
    categoryID := c.Param("id")

    var articles []Article
    db.Where("category_id = ?", categoryID).Find(&articles)

    c.JSON(http.StatusOK, articles)
}

// @Summary 点赞文章
// @Description 给指定文章点赞，需要JWT认证
// @Tags 点赞
// @Param Authorization header string true "Bearer token"
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /articles/{id}/like [post]
func likeArticle(c *gin.Context) {
    articleID := c.Param("id")

    // 获取当前用户
    username, _ := c.Get("username")
    var user User
    if db.Where("username = ?", username).First(&user).Error != nil {
        c.JSON(404, gin.H{"error": "用户不存在"})
        return
    }

    // 检查文章存在
    var article Article
    if db.First(&article, articleID).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    // 检查是否已点赞
    var existing Like
    if db.Where("user_id = ? AND article_id = ?", user.ID, articleID).First(&existing).Error == nil {
        c.JSON(400, gin.H{"error": "已点赞"})
        return
    }

    like := Like{UserID: user.ID, ArticleID: article.ID}
    db.Create(&like)

    // 统计点赞数
    count := db.Model(&article).Association("Likes").Count()

    c.JSON(200, gin.H{"message": "点赞成功", "likes_count": count})
}

// @Summary 取消点赞
// @Description 取消对指定文章的点赞
// @Tags 点赞
// @Param Authorization header string true "Bearer token"
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /articles/{id}/like [delete]
func unlikeArticle(c *gin.Context) {
    articleID := c.Param("id")

    username, _ := c.Get("username")
    var user User
    db.Where("username = ?", username).First(&user)

    var like Like
    if db.Where("user_id = ? AND article_id = ?", user.ID, articleID).First(&like).Error != nil {
        c.JSON(404, gin.H{"error": "未点赞"})
        return
    }

    db.Delete(&like)

    var article Article
    db.First(&article, articleID)
    count := db.Model(&article).Association("Likes").Count()

    c.JSON(200, gin.H{"message": "取消点赞成功", "likes_count": count})
}

// @Summary 获取点赞数
// @Description 获取指定文章的点赞数
// @Tags 点赞
// @Param id path int true "文章ID"
// @Success 200 {object} map[string]interface{}
// @Router /articles/{id}/likes [get]
func getLikesCount(c *gin.Context) {
    articleID := c.Param("id")

    var article Article
    if db.First(&article, articleID).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    count := db.Model(&article).Association("Likes").Count()

    c.JSON(200, gin.H{"article_id": articleID, "likes_count": count})
}

// ========== 主函数 ==========

func main() {
    // 初始化数据库
    var err error
    db, err = gorm.Open(sqlite.Open("./data/gin_blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&User{}, &Article{}, &Comment{}, &Category{}, &Like{})

    r := gin.Default()

    // Swagger文档路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    r.Use(func(c *gin.Context) {
        start := time.Now()
        c.Next()

        fmt.Printf("[%s] %s %s | %d | %v\n",
            time.Now().Format("2006-01-02 15:04:05"),
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            time.Since(start),
        )
    })

    // 公开路由
    r.POST("/register", register)
    r.POST("/login", login)
    r.GET("/article/:id/comments", getArticleComments)
    r.GET("/categories", listCategories)
    r.GET("/categories/:id/articles", getCategoryArticles)
    r.GET("/articles/:id/likes", getLikesCount)

    //需要认证的路由
    auth := r.Group("/")
    auth.Use(JWTAuth())
    {
        auth.GET("/me", getMe)
        auth.PUT("/users/me", updateProfile)
        auth.POST("/articles", createArticle)
        auth.PUT("/articles/:id", updateArticle)
        auth.DELETE("/articles/:id", deleteArticle)
        auth.POST("/comments", createComment)
        auth.POST("/categories", createCategory)
        auth.PUT("/categories/:id", updateCategory)
        auth.DELETE("/categories/:id", deleteCategory)
        auth.POST("/articles/:id/like", likeArticle)
        auth.DELETE("/articles/:id/like", unlikeArticle)
    }

    r.Run(":8080")
}
