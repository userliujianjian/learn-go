package example

import "fmt"

// SliceCapAddr 验证切片容量增长后底层元素是否还有依赖
func SliceCapAddr() {
	fmt.Println("[SliceCapAddr] is start...")
	num := make([]int, 3, 4)
	num1 := append(num, 1)
	num2 := append(num1, 2)
	num1[0] = 1
	num2[1] = 2
	fmt.Printf("[SliceCapAddr] num: %v, cap: %d \n", num, cap(num))    // [SliceCapAddr] num: [1 0 0], cap: 4
	fmt.Printf("[SliceCapAddr] num1: %v, cap: %d \n", num1, cap(num1)) // [SliceCapAddr] num1: [1 0 0 1], cap: 4
	fmt.Printf("[SliceCapAddr] num2: %v, cap: %d \n", num2, cap(num2)) // [SliceCapAddr] num2: [0 2 0 1 2], cap: 8

	fmt.Println("[SliceCapAddr] is end...")
}

func SliceCap() {
	fmt.Println("[SliceCap] is start...")
	s := make([]int, 0)
	oldCap := cap(s)

	for i := 0; i < 2048; i++ {
		s = append(s, i)

		newCap := cap(s)

		if newCap != oldCap {
			fmt.Printf("[SliceCap][%d -> %4d] cap = %-4d  |  after append %-4d  cap = %-4d\n", 0, i-1, oldCap, i, newCap)
			oldCap = newCap
		}
	}
	fmt.Println("[SliceCap] is end...")

	//[SliceCap] is start...
	//[SliceCap][0 ->   -1] cap = 0     |  after append 0     cap = 1
	//[SliceCap][0 ->    0] cap = 1     |  after append 1     cap = 2
	//[SliceCap][0 ->    1] cap = 2     |  after append 2     cap = 4
	//[SliceCap][0 ->    3] cap = 4     |  after append 4     cap = 8
	//[SliceCap][0 ->    7] cap = 8     |  after append 8     cap = 16
	//[SliceCap][0 ->   15] cap = 16    |  after append 16    cap = 32
	//[SliceCap][0 ->   31] cap = 32    |  after append 32    cap = 64
	//[SliceCap][0 ->   63] cap = 64    |  after append 64    cap = 128
	//[SliceCap][0 ->  127] cap = 128   |  after append 128   cap = 256
	//[SliceCap][0 ->  255] cap = 256   |  after append 256   cap = 512
	//[SliceCap][0 ->  511] cap = 512   |  after append 512   cap = 848
	//[SliceCap][0 ->  847] cap = 848   |  after append 848   cap = 1280
	//[SliceCap][0 -> 1279] cap = 1280  |  after append 1280  cap = 1792
	//[SliceCap][0 -> 1791] cap = 1792  |  after append 1792  cap = 2560
	//[SliceCap] is end...
}

func SliceCapPlug1() {
	// s只有一个元素[5]
	s := []int{5}
	// s扩容，容量变为2，[5,7]
	s = append(s, 7)

	// s扩容，容量变为4 【5，7，9】，注意，这时s长度为3，只有3个元素
	s = append(s, 9)

	// 由于s的底层数组仍然有空间，并不会扩容。
	// 这样底层数组变成[5,7,9,11], 注意此时 s = [5,7,9], 容量为4；
	// x = [5,7,9,11] 容量为4，这里s不变
	x := append(s, 11)

	// 这里还是s元素的尾部追加元素，由于s的长度为3，容量为4，
	// 所以直接在底层数组所因为3的地方填上12，结果 s = [5,7,9], y=[5,7,9,12], x=[5,7,9,12]
	// x和y的长度均为4，容量也均为4
	y := append(s, 12)
	fmt.Println("[SliceCapPlug1]", s, x, y)
}

func SliceCapPlug2() {
	s := []int{1, 2}
	s = append(s, 3, 4, 5)
	fmt.Printf("[SliceCapPlug2] len=%d, cap=%d \n", len(s), cap(s))

}

func SliceCapPlug() {
	SliceCapPlug1()
	SliceCapPlug2()
}
