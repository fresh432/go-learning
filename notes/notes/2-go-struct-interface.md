# Go 结构体、方法、接口学习笔记

## 学习时间
2026-07-01 15:10-17:40（约2.5小时）

## 核心概念

### 1. 结构体
```go
type Person struct {
    Name string
    Age  int
}
```

### 2. 方法
```go
// 值接收者（不修改原对象）
func (p Person) GetInfo() string {}

// 指针接收者（可修改原对象）
func (p *Person) HaveBirthday() {}
```

### 3. 接口（隐式实现）
```go
type Geometry interface {
    Area() float64
    Perimeter() float64
}

// 只要实现了Area()和Perimeter()，就自动实现Geometry接口
// 不需要显式声明 `implements`
```

## 与Python对比
| Python  | Go                  |
| ------- | ------------------- |
| `class` | `struct` + `method` |
| 显式继承    | 隐式实现接口              |
| `self`  | 显式接收者参数             |


## 关键
- Go没有class，用struct+method模拟
- 接口隐式实现，更灵活
- 指针接收者 vs 值接收者要区分清楚

## 下一步
并发、HTTP服务