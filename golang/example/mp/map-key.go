package main

import "fmt"

func main() {
	ageMp := make(map[string]int)
	// 指定map长度
	ageMp2 := make(map[string]int, 8)

	// ageMp3 为nil，不能向其添加元素，会直接panic
	var ageMp3 map[string]int

	fmt.Println(ageMp, ageMp2, ageMp3)
}
