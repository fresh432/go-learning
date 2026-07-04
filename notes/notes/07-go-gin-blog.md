# Go Gin完整博客API项目笔记

## 学习时间
2026-07-04 16:43-18:20（约1.5小时）

## 项目结构
07-gin-blog.go
├── 模型定义（Article + Category）
├── 数据库初始化（GORM + SQLite）
├── Handler函数（CRUD）
├── 路由注册
└── 主函数


## 核心代码解析

### 1. 模型定义
```go
type Article struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    Title      string    `json:"title" gorm:"not null"`
    Content    string    `json:"content"`
    Author     string    `json:"author" gorm:"default:'匿名'"`
    CategoryID *uint     `json:"category_id"`  // 指针，可为NULL
    CreatedAt  time.Time `json:"created_at"`
}

type Category struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Name string `json:"name" gorm:"unique;not null"`
}
```

#### 与FastAPI对比：
| FastAPI                               | Gin                   |
| ------------------------------------- | --------------------- |
| `Column(Integer, primary_key=True)`   | `gorm:"primaryKey"`   |
| `Column(String(100), nullable=False)` | `gorm:"not null"`     |
| `ForeignKey("categories.id")`         | `CategoryID *uint` 指针 |
| `relationship("Category")`            | 无自动关联，手动查询            |

### 2. 数据库初始化
```go
func initDB() {
    db, err = gorm.Open(sqlite.Open("gin_blog.db"), &gorm.Config{})
    db.AutoMigrate(&Article{}, &Category{})
}
```

### 3. CRUD Handler
| 操作   | 代码                                 | 对比FastAPI                               |
| ---- | ---------------------------------- | --------------------------------------- |
| 查询所有 | `db.Find(&articles)`               | `db.query(Article).all()`               |
| 查询单条 | `db.First(&article, id)`           | `db.query(Article).filter(...).first()` |
| 创建   | `db.Create(&article)`              | `db.add(article)` + `db.commit()`       |
| 更新   | `db.Model(&article).Updates(data)` | 修改属性 + `db.commit()`                    |
| 删除   | `db.Delete(&Article{}, id)`        | `db.delete(article)` + `db.commit()`    |

### 4. 路由注册
```go
r.GET("/articles", listArticles)
r.GET("/articles/:id", getArticle)
r.POST("/articles", createArticle)
r.PUT("/articles/:id", updateArticle)
r.DELETE("/articles/:id", deleteArticle)

r.GET("/categories", listCategories)
r.POST("/categories", createCategory)
r.GET("/categories/:id/articles", getCategoryArticles)
```

## 踩坑记录
| 问题              | 解决                                                |
| --------------- | ------------------------------------------------- |
| CategoryID为NULL | 用指针 `*uint`，json中为null                            |
| 更新操作            | `db.Model(&article).Updates(data)` 不是 `db.Update` |
| 路由参数            | `:id` 不是 `{id}`                                   |
| 状态码             | 需手动设置 `http.StatusOK`                             |

## 关键
- Go开发比Python更"底层"，需手动处理更多细节
- GORM功能强大，但语法与SQLAlchemy有差异
- Gin性能高，但开发效率不如FastAPI
- 适合：高性能场景、微服务、网关

## 下一步
- 添加JWT认证（对比FastAPI实现）
- 添加用户系统
- 部署到Docker
- 与Python项目联调