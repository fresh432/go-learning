# Go Gin博客API最终版（10-gin-final.go）

## 学习时间
2026-07-14 16:08-18:03（约2小时）

## 本次新增

### 1. 文章更新（PUT）

```go
func updateArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if db.First(&article, id).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    var updateData Article
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 验证分类存在
    if updateData.CategoryID != nil && *updateData.CategoryID != 0 {
        var category Category
        if db.First(&category, *updateData.CategoryID).Error != nil {
            c.JSON(404, gin.H{"error": "分类不存在"})
            return
        }
    }

    // 保留原作者
    updateData.Author = article.Author

    db.Model(&article).Updates(updateData)
    c.JSON(200, article)
}
```

#### 关键:
- Updates() 只更新非零值字段
- 保留原作者，防止篡改
- 分类ID为指针类型，需判断非空

### 2. 文章删除（DELETE）
```go
func deleteArticle(c *gin.Context) {
    id := c.Param("id")
    var article Article
    if db.First(&article, id).Error != nil {
        c.JSON(404, gin.H{"error": "文章不存在"})
        return
    }

    // 手动删除关联评论
    db.Where("article_id = ?", id).Delete(&Comment{})

    db.Delete(&article)
    c.JSON(200, gin.H{"message": "删除成功"})
}
```

#### 关键:
- 先查再删，确认存在
- 手动清理关联评论，避免孤儿数据

### 3. 分类更新（PUT）
```go
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
```

### 4. 路由注册
```go
auth.PUT("/articles/:id", updateArticle)
auth.DELETE("/articles/:id", deleteArticle)
auth.PUT("/categories/:id", updateCategory)
auth.DELETE("/categories/:id", deleteCategory)
```

## 踩坑记录
| 问题            | 原因             | 解决          |
| ------------- | -------------- | ----------- |
| Updates只更新非零值 | GORM特性，零值字段被忽略 | 用指针类型或单独处理  |
| 删除关联数据        | 外键约束可能报错       | 手动先删评论，再删文章 |
| 作者被覆盖         | 更新时传了author字段  | 强制保留原author |

## 下一步
- 学习Go部署（Docker交叉编译）
- 性能测试：Go vs Python
- 项目文档完善
