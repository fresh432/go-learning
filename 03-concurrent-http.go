/*
Go 并发与HTTP服务练习
学习时间：2026-07-01 19:40-21:09
来源：Go by Example
*/

package main

import (
	"fmt"
	"net/http"
	"time"
)

// ========== Goroutine（轻量级线程）==========
func say(s string) {
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

// ========== Channel（协程间通信）==========
func sum(a []int, c chan int) {
	total := 0
	for _, v := range a {
		total += v
	}
	c <- total  // 发送结果到channel
}

// ========== HTTP Handler ==========
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from Go! 时间: %s\n", time.Now().Format("15:04:05"))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Go HTTP服务运行中\n")
}

func main() {
	// ========== Goroutine 示例 ==========
	fmt.Println("=== Goroutine ===")
	go say("world")  // 新开协程
	say("hello")     // 主协程

	// ========== Channel 示例 ==========
	fmt.Println("\n=== Channel ===")
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8}
	c := make(chan int)

	go sum(nums[:len(nums)/2], c)  // 前半部分
	go sum(nums[len(nums)/2:], c)  // 后半部分

	x, y := <-c, <-c  // 接收两个结果
	fmt.Printf("x=%d, y=%d, total=%d\n", x, y, x+y)

	// ========== HTTP 服务 ==========
	fmt.Println("\n=== HTTP服务 ===")
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/info", infoHandler)

	fmt.Println("服务启动: http://localhost:8090")
	fmt.Println("访问: /hello 或 /info")
	http.ListenAndServe(":8090", nil)
}