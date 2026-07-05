# Go Gin JWT认证学习笔记

## 学习时间
2026-07-05 14:38-17:06（约2.5小时）

## 核心内容

### 1. JWT库选择

| 语言 | 库 |
|------|-----|
| Python | `python-jose` |
| Go | `github.com/golang-jwt/jwt/v5` |

**安装：**
```bash
go get -u github.com/golang-jwt/jwt/v5
go get -u golang.org/x/crypto/bcrypt
```

### 2. Token生成与解析
#### 生成Token
```go
type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

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
```

#### 与Python对比:
| Python                                | Go                                                           |
| ------------------------------------- | ------------------------------------------------------------ |
| `jwt.encode(data, secret, algorithm)` | `jwt.NewWithClaims(method, claims)` + `SignedString(secret)` |

#### 解析Token
```go
func parseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, err
}
```

#### 与Python对比：
| Python                                        | Go                                               |
| --------------------------------------------- | ------------------------------------------------ |
| `jwt.decode(token, secret, algorithms=[...])` | `jwt.ParseWithClaims(token, &Claims{}, keyFunc)` |

### 3. 密码加密
| 操作 | Python (passlib)             | Go (bcrypt)                                                |
| -- | ---------------------------- | ---------------------------------------------------------- |
| 加密 | `get_password_hash(pwd)`     | `bcrypt.GenerateFromPassword([]byte(pwd), 14)`             |
| 验证 | `verify_password(pwd, hash)` | `bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))` |

### 4. 认证中间件
```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        
        // 去掉 "Bearer " 前缀
        tokenString := authHeader
        if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
            tokenString = authHeader[7:]
        }
        
        claims, err := parseToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "无效token"})
            c.Abort()
            return
        }
        
        // 存入上下文
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

#### 与FastAPI对比：
| 特性      | FastAPI                     | Gin                               |
| ------- | --------------------------- | --------------------------------- |
| 获取Token | `OAuth2PasswordBearer` 自动   | 手动 `c.GetHeader("Authorization")` |
| 解析Token | `Depends(get_current_user)` | 自定义中间件                            |
| 传递用户信息  | 函数参数注入                      | `c.Set("key", value)` 上下文         |

### 5. 上下文传递
```go
// 中间件中存入
c.Set("username", claims.Username)

// Handler中获取
username, _ := c.Get("username")
article.Author = username.(string)
```

#### 关键:
Go是静态类型，需要类型断言 .(string)

### 6. 路由分组
```go
// 公开路由
r.POST("/register", register)
r.POST("/login", login)

// 需要认证
auth := r.Group("/")
auth.Use(JWTAuth())
{
    auth.GET("/me", getMe)
    auth.POST("/articles", createArticle)
}
```

### 7. 完整流程对比
| 步骤         | FastAPI                     | Gin                         |
| ---------- | --------------------------- | --------------------------- |
| 1. 注册      | `POST /register`            | `POST /register`            |
| 2. 登录      | `POST /token`               | `POST /login`               |
| 3. 获取Token | 返回 `access_token`           | 返回 `access_token`           |
| 4. 访问资源    | `Authorization: Bearer xxx` | `Authorization: Bearer xxx` |
| 5. 获取用户    | `Depends(get_current_user)` | `c.Get("username")`         |

## 踩坑记录
| 问题             | 解决                           |
| -------------- | ---------------------------- |
| `Bearer ` 前缀处理 | 手动截取 `authHeader[7:]`        |
| 类型断言           | `username.(string)`          |
| Token过期        | `RegisteredClaims.ExpiresAt` |
| 密码字段隐藏         | `json:"-"` tag               |

## 关键
- Go JWT处理比Python更底层，需手动处理更多细节
- 上下文传递是Go Web开发核心模式
- 类型安全是Go优势，但也增加代码量
- 理解原理后，两种语言实现逻辑相同

## 下一步
- Go项目部署（Docker）
- 与Python项目联调
- 学习Go并发模式在Web中的应用
