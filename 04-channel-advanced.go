/*
Go Channel进阶练习
学习时间：2026-07-02 20:06-20:35
来源：Go by Example
*/

package main

import (
	"fmt"
	"time"
)

func main() {
	// ========== Select：多Channel选择 ==========
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "来自ch1"
	}()

	go func() {
		time.Sleep(2 * time.Second)
		ch2 <- "来自ch2"
	}()

	// 等待两个channel，先到达的先执行
	select {
	case msg1 := <-ch1:
		fmt.Println(msg1)
	case msg2 := <-ch2:
		fmt.Println(msg2)
	}

	// ========== Timeout：超时控制 ==========
	ch := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- "结果"
	}()

	select {
	case res := <-ch:
		fmt.Println(res)
	case <-time.After(1 * time.Second):  // 1秒超时
		fmt.Println("超时！")
	}

	// ========== Non-Blocking：非阻塞操作 ==========
	messages := make(chan string)
	signals := make(chan bool)

	// 非阻塞接收
	select {
	case msg := <-messages:
		fmt.Println("收到:", msg)
	default:
		fmt.Println("无消息")  // 不阻塞，直接执行
	}

	// 非阻塞发送
	select {
	case messages <- "hi":
		fmt.Println("发送成功")
	default:
		fmt.Println("发送失败，channel满")
	}

	// ========== Close：关闭Channel ==========
	jobs := make(chan int, 5)
	for j := 1; j <= 3; j++ {
		jobs <- j
	}
	close(jobs)  // 关闭后不能再发送，但可以接收

	// Range遍历已关闭的channel
	for job := range jobs {
		fmt.Println("处理任务:", job)
	}
}