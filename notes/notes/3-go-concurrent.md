# Go 并发与HTTP服务学习笔记

## 学习时间
2026-07-01 19:40-21:09（约1.5小时）

## 核心概念

### 1. Goroutine（协程）
```go
go say("world")  // 新开协程，不阻塞
```

### 2. Channel（通信）
```go
c := make(chan int)
c <- value      // 发送
value := <-c    // 接收
```

### 3. HTTP服务
```go
http.HandleFunc("/hello", handler)
http.ListenAndServe(":8090", nil)
```

## 与Python对比
| Python         | Go                      |
| -------------- | ----------------------- |
| `threading`（重） | `goroutine`（轻，万级）       |
| `asyncio`      | `goroutine` + `channel` |
| Flask/FastAPI  | 内置 `net/http`           |


## 关键
- Goroutine极轻量，可开成千上万个
- Channel是协程间通信方式，不要共享内存
- Go内置HTTP服务，不需要额外框架

## 下一步
- 深入Channel、Select
- 学习Gin框架（Web开发）
- 与Python FastAPI项目联调