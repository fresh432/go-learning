package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

    "gorm.io/gorm"
	sqlite "github.com/ncruces/go-sqlite3/gormlite"
)

// 初始化测试数据库
func setupTestDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("./data/test_gin_blog.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}
	// 自动迁移
	db.AutoMigrate(&User{}, &Article{}, &Comment{}, &Category{}, &Like{})
}

// 获取测试用的gin引擎
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 注册公开路由
	r.POST("/register", register)
	r.POST("/login", login)
	r.GET("/categories", listCategories)
	r.GET("/articles/:id/likes", getLikesCount)

	// 认证路由
	auth := r.Group("/")
	auth.Use(JWTAuth())
	{
		auth.GET("/me", getMe)
		auth.PUT("/users/me", updateProfile)
		auth.POST("/articles", createArticle)
		auth.POST("/articles/:id/like", likeArticle)
		auth.DELETE("/articles/:id/like", unlikeArticle)
	}

	return r
}

// ========== 测试用例 ==========

func TestRegister(t *testing.T) {
	setupTestDB()
	r := setupRouter()

    db.Where("username = ?", "testuser").Delete(&User{})

	// 注册新用户
	body := `{"username":"testuser","password":"123456"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	assert.Contains(t, w.Body.String(), "注册成功")
}

func TestLogin(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	// 先注册
	body := `{"username":"logintest","password":"123456"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 再登录
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
}

func TestListCategories(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/categories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetMeUnauthorized(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/me", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "未授权")
}

func TestCreateArticleUnauthorized(t *testing.T) {
	setupTestDB()
	r := setupRouter()

	body := `{"title":"Test","content":"Test content"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/articles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}