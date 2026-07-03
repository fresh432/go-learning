# Go Gin框架深入学习笔记

## 学习时间
2026-07-03 15:14-17:34（约2.5小时）

## 核心内容

### 1. 中间件（Middleware）

#### 日志中间件
```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()  // 执行后续handler
        // 记录请求信息
        fmt.Printf("%s %s | 状态:%d | 耗时:%v\n",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            time.Since(start))
    }
}
```

#### 认证中间件
```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "未授权"})
            c.Abort()  // 终止后续处理
            return
        }
        c.Next()
    }
}
```

#### 关键
- c.Next() 执行后续handler
- c.Abort() 终止请求链
- r.Use(Logger()) 全局注册

### 2. 路由分组 （Router Group）
```go
v1 := r.Group("/api/v1")
{
    // 文章路由
    articles := v1.Group("/articles")
    {
        articles.GET("", listArticles)
        articles.GET("/:id", getArticle)
        articles.POST("", createArticle)
    }
    
    // 需要认证
    auth := v1.Group("/")
    auth.Use(Auth())
    {
        auth.GET("/me", getMe)
    }
}
```

#### 与FastAPI对比：
| 特性   | FastAPI                       | Gin                  |
| ---- | ----------------------------- | -------------------- |
| 路由分组 | `APIRouter(prefix="/api/v1")` | `r.Group("/api/v1")` |
| 依赖注入 | `Depends(get_db)`             | 无原生依赖注入，用中间件         |
| 认证   | `OAuth2PasswordBearer`        | 自定义中间件               |

### 3. GORM数据库操作
#### 连接数据库
```go
import (
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)

db, err := gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
db.AutoMigrate(&Article{})
```
#### CRUD操作
| 操作   | GORM                        | SQLAlchemy                              |
| ---- | --------------------------- | --------------------------------------- |
| 查询所有 | `db.Find(&articles)`        | `db.query(Article).all()`               |
| 查询单条 | `db.First(&article, id)`    | `db.query(Article).filter(...).first()` |
| 创建   | `db.Create(&article)`       | `db.add(article)` + `db.commit()`       |
| 删除   | `db.Delete(&Article{}, id)` | `db.delete(article)` + `db.commit()`    |

#### 模型定义
```go
type Article struct {
    ID      uint   `json:"id" gorm:"primaryKey"`
    Title   string `json:"title" gorm:"not null"`
    Content string `json:"content"`
}
```

#### GORM tag 对应 SQLAlchemy：
- gorm:"primaryKey" → primary_key=True
- gorm:"not null" → nullable=False
- gorm:"default:'匿名'" → default="匿名"

### 4. 完整CRUD示例
```go
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
```

## 与FastAPI核心差异
| 维度   | FastAPI                  | Gin          |
| ---- | ------------------------ | ------------ |
| 开发效率 | **高**（自动生成文档、Pydantic验证） | 中（手动处理多）     |
| 性能   | 快                        | **更快**       |
| 异步   | 原生支持                     | 需配合goroutine |
| 依赖注入 | 原生支持                     | 无，用全局变量或中间件  |
| 适用场景 | 快速开发、API服务               | 高性能、微服务、网关   |

## 下一步
- 用Gin写完整博客API（路由、模型、数据库）
- 学习Gin部署（Docker、反向代理）
- 与Python FastAPI项目对比联调
