/*
Go 结构体、方法、接口练习
学习时间：2026-07-01 15:10-17:40
来源：Go by Example
*/

package main

import (
	"fmt"
	"math"
)

// ========== 结构体 ==========
type Person struct {
	Name string
	Age  int
}

// ========== 方法（值接收者）==========
func (p Person) GetInfo() string {
	return fmt.Sprintf("%s, %d岁", p.Name, p.Age)
}

// ========== 方法（指针接收者，可修改）==========
func (p *Person) HaveBirthday() {
	p.Age++
}

// ========== 接口 ==========
type Geometry interface {
	Area() float64
	Perimeter() float64
}

// ========== 实现接口：矩形 ==========
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// ========== 实现接口：圆 ==========
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// ========== 接口使用 ==========
func measure(g Geometry) {
	fmt.Printf("面积: %.2f, 周长: %.2f\n", g.Area(), g.Perimeter())
}

func main() {
	// 结构体
	p := Person{Name: "张三", Age: 20}
	fmt.Println(p.GetInfo())

	p.HaveBirthday()  // 指针接收者，Age变为21
	fmt.Println(p.GetInfo())

	// 接口
	r := Rectangle{Width: 3, Height: 4}
	c := Circle{Radius: 5}

	measure(r)  // 矩形
	measure(c)  // 圆

	// 空接口（任意类型）
	var any interface{} = "hello"
	fmt.Println(any)
}