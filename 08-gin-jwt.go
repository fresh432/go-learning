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
    ID          uint    `json:"id" gorm:"primaryKey"`
    Title       string  `json:"title" gorm:"not null"`
    Content     string  `json:"content"`
    Author      string  `json:"author"`
    CategoryID  *uint   `json:"category_id"`
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

// ========== 主函数 ==========

func main() {
    // 初始化数据库
    var err error
    db, err = gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    db.AutoMigrate(&User{}, &Article{})

    r := gin.Default()

    // 公开路由
    r.POST("/register", register)
    r.POST("/login", login)

    //需要认证的路由
    auth := r.Group("/")
    auth.Use(JWTAuth())
    {
        auth.GET("/me", getMe)
        auth.POST("/articles", createArticle)
    }

    r.Run(":8080")
}
