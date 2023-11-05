package example

import "fmt"

func SliceVar() {
	var nums []int                    // 生命切片
	fmt.Println(len(nums), cap(nums)) // 0, 0
	nums = append(nums, 1)            // 初始化
	fmt.Println(len(nums), cap(nums)) // 1, 1

	nums1 := []int{1, 2, 3, 4}          // 生命并初始化
	fmt.Println(len(nums1), cap(nums1)) // 4 4

	nums2 := make([]int, 3, 5)          // 使用make函数构造切片
	fmt.Println(len(nums2), cap(nums2)) // 3 5

}

func SliceFunc(numbers []int) {
	for i := 0; i < len(numbers); i++ {
		numbers[i] = numbers[i] + 1
	}
	fmt.Println("numbers is SliceFunc: ", numbers) // [2 3 4 5 6]
}

func SliceMain() {
	var numbers []int
	for i := 0; i < 5; i++ {
		numbers = append(numbers, i+1)
	}
	SliceFunc(numbers)
	// 切片被当作参数传递时，直接修改的是切片本身
	fmt.Println("numbers in main: ", numbers) // [2 3 4 5 6]
}
