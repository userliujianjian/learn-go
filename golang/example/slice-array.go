package example

import "fmt"

// SliceTest 切片测试
func SliceTest() {
	slice := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	s1 := slice[2:5] // 从下标2开始，到下标5结束，不包含5
	// 从s1的下标2开始，到下标6结束，不包含6， 容量到下标7
	//  slice[4:8:9] 与s2有什么区别
	s2 := s1[2:6:7]

	s2 = append(s2, 100)
	s2 = append(s2, 200)

	s1[2] = 20
	fmt.Println("SliceTest start.....")

	fmt.Println(s1)
	fmt.Println(s2)
	fmt.Println(slice)
	fmt.Println("SliceTest end.....")
}
