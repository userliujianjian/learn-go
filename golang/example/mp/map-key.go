package main

import (
	"fmt"
	"math"
)

func main() {
	ageMp := make(map[string]int)
	// 指定map长度
	ageMp2 := make(map[string]int, 8)

	// ageMp3 为nil，不能向其添加元素，会直接panic
	var ageMp3 map[string]int

	fmt.Println(ageMp, ageMp2, ageMp3)
	GetKey()
	floatKey()
}

func GetKey() {
	ageMap := make(map[string]int)
	ageMap["hll"] = 18
	// 不带comma用法
	age1 := ageMap["lhh"]
	fmt.Println(age1) // 0

	// 带comma 用法
	age2, ok := ageMap["lhh"]
	fmt.Println(age2, ok) // 0 false
}

// 浮点数可以作为key？
func floatKey() {
	m := make(map[float64]int)

	m[1.4] = 1
	m[2.4] = 2
	m[math.NaN()] = 3
	m[math.NaN()] = 3

	for k, v := range m {
		fmt.Printf("[%v , %d] \n", k, v)
	}

	fmt.Printf("k: %v, v: %d \n", math.NaN(), m[math.NaN()])                                   // k: NaN, v: 0
	fmt.Printf("k: %v, v: %d \n", 2.400000000001, m[2.400000000001])                           // k: 2.400000000001, v: 0
	fmt.Printf("k: %v, v: %d \n", 2.4000000000000000000000001, m[2.4000000000000000000000001]) // k: 2.4, v: 2
}
