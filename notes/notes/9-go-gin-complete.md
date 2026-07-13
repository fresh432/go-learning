# Go Gin完整博客API（09-gin-complete.go）补充笔记
 
## 学习时间
2026-07-13 14:40-17:00（约2.5小时）

### 1. 新增功能
- 评论系统 Comment模型+创建+获取 
- 分类关联 Category外键关联 
- 文章-评论一对多 `Comments []Comment` 
- 评论作者自动填充 从JWT上下文获取 

### 2. 模型关联写法

```go
// Article添加评论关联
type Article struct {
    // ... 原有字段
    Comments []Comment `json:"comments" gorm:"foreignKey:ArticleID"`
}

// Comment模型
type Comment struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Content   string    `json:"content" gorm:"not null"`
    Author    string    `json:"author"`
    ArticleID uint      `json:"article_id" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`  // 注意：At不是AT
}
```

### 3. 评论创建流程
```go
func createComment(c *gin.Context) {
    var comment Comment
    c.ShouldBindJSON(&comment)
    
    // 1. 验证文章存在
    var article Article
    db.First(&article, comment.ArticleID)
    
    // 2. 从JWT上下文获取当前用户
    username, _ := c.Get("username")
    comment.Author = username.(string)
    
    // 3. 创建
    db.Create(&comment)
}
```

#### 与FastAPI对比：
| 步骤     | FastAPI                                 | Go Gin                   |
| ------ | --------------------------------------- | ------------------------ |
| 验证文章   | `db.query(Article).filter(...).first()` | `db.First(&article, id)` |
| 获取当前用户 | `Depends(get_current_user)`             | `c.Get("username")`      |
| 类型处理   | 动态类型                                    | `username.(string)` 类型断言 |


### 4. 与FastAPI项目功能对齐表
| 功能      | FastAPI   | Go Gin（09）    | 差异            |
| ------- | --------- | ------------- | ------------- |
| 用户注册/登录 | ✅         | ✅             | 无             |
| JWT认证   | ✅         | ✅             | 无             |
| 密码加密    | ✅ bcrypt  | ✅ bcrypt      | 无             |
| 文章创建    | ✅         | ✅             | 无             |
| 文章更新/删除 | ✅         | ❌             | Go缺PUT/DELETE |
| 分类CRUD  | ✅ 完整      | ✅ 缺PUT/DELETE | 同上            |
| 评论功能    | ✅         | ✅             | 无             |
| 获取当前用户  | ✅         | ✅             | 无             |
| 自动API文档 | ✅ Swagger | ❌             | Go需额外配置       |

## 下一步
- 添加文章/分类的PUT/DELETE，完全闭环
- 学习Go部署（Docker交叉编译）
- 性能测试：Go vs Python
