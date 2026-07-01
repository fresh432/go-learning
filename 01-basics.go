/*
Go 基础语法练习
学习时间：2026-07-01 13:40-15:10
来源：Go by Example
*/

package main

import (
	"fmt"
)

func main() {
	// ========== 变量与常量 ==========
	var a int = 10          // 显式声明
	b := 20                 // 类型推断
	const pi = 3.14159      // 常量

	fmt.Printf("a=%d, b=%d, pi=%f\n", a, b, pi)

	// ========== 数组（固定长度）==========
	var arr [5]int = [5]int{1, 2, 3, 4, 5}
	fmt.Println("数组:", arr)

	// ========== 切片（动态数组，Go核心）==========
	s := make([]int, 3)     // 长度3，容量3
	s = append(s, 4)        // 追加元素
	s = append(s, 5, 6)     // 追加多个
	fmt.Println("切片:", s)

	// 切片截取
	sub := s[1:4]           // 索引1到3
	fmt.Println("截取:", sub)

	// ========== map（哈希表）==========
	m := make(map[string]int)
	m["age"] = 20
	m["score"] = 90

	// 取值，判断是否存在
	if v, ok := m["age"]; ok {
		fmt.Printf("age=%d\n", v)
	}

	// 删除
	delete(m, "score")

	// 遍历
	for k, v := range m {
		fmt.Printf("%s: %d\n", k, v)
	}

	// ========== 函数多返回值 ==========
	sum, diff := calc(10, 3)
	fmt.Printf("sum=%d, diff=%d\n", sum, diff)

	// 忽略返回值
	sum2, _ := calc(5, 2)
	fmt.Println("sum2:", sum2)
}

// 多返回值函数
func calc(x, y int) (int, int) {
	return x + y, x - y
}