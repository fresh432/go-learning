# Go Channel进阶学习笔记

## 学习时间
2026-07-02 20:06-20:40（约40分钟）

## 核心概念

### 1. Select：多Channel选择
```go
select {
case msg1 := &lt;-ch1:
    // ch1有数据
case msg2 := &lt;-ch2:
    // ch2有数据
}
```

- 类似switch，但用于channel
- 随机选择可用的case
- 可用于超时、非阻塞操作

### 2. Timeout：超时控制
```go
select {
case res := <-ch:
    // 正常接收
case <-time.After(1 * time.Second):
    // 超时处理
}
```

### 3. Non-Blocking：非阻塞
```go
select {
case msg := <-ch:
    // 有数据
default:
    // 无数据，不阻塞
}
```

### 4. Close：关闭Channel
- close(ch) 关闭后不能再发送
- 可以接收已关闭channel的剩余数据
- range 遍历会自动处理关闭

## 下一步
Gin框架入门，写Web API