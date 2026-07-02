# Go Gin框架入门笔记

## 学习时间
2026-07-02 22:40-22:10（约90分钟）

## 核心概念

### 1. Gin是什么
- Go语言最流行的Web框架
- 类似Python的Flask，比FastAPI轻量
- 性能极高，路由简单

### 2. 基本路由
```go
r := gin.Default()
r.GET("/path", handler)
r.POST("/path", handler)
```

### 3. 参数获取
| 方式        | 代码                       |
| --------- | ------------------------ |
| URL参数     | `c.Param("id")`          |
| Query参数   | `c.Query("key")`         |
| JSON Body | `c.ShouldBindJSON(&obj)` |
| Header    | `c.GetHeader("key")`     |

### 4. 响应
```go
c.JSON(200, gin.H{"key": "value"})  // JSON
c.String(200, "hello")               // 字符串
```

## 与FastAPI对比
| 特性   | FastAPI    | Gin          |
| ---- | ---------- | ------------ |
| 语言   | Python     | Go           |
| 异步   | 原生支持       | 需配合goroutine |
| 参数验证 | Pydantic自动 | 需手动绑定        |
| 文档   | 自动生成       | 需额外配置        |
| 性能   | 快          | **更快**       |
| 开发效率 | **高**      | 中            |

## 关键
- in性能极高，但开发效率不如FastAPI
- Go Web开发需要手动处理更多细节
- 适合高性能场景，不适合快速原型

## 下一步
- 学习Gin中间件、路由分组
- 与FastAPI项目对比，理解差异
- 尝试用Gin重写部分API