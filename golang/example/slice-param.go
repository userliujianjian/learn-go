package example

import "fmt"

func SliceParam() {
	exp1()
	exp2()
}

// exp1 切片作为参数，改变副本底层数据测试
func exp1() {
	s := []int{1, 2, 3}
	SliceAdd(s)
	fmt.Printf("[exp1] 原始数据 s: %v, cap: %d \n", s, cap(s))    //[SliceParamExp1] 原始数据 s: [2 3 4], cap: 3
	fmt.Printf("[exp1] SliceAdd s: %v, cap: %d\n", s, cap(s)) //[SliceParamExp1] SliceAdd s: [2 3 4], cap: 3
}

// exp2 切片作为参数，改变副本底层数据测试2
func exp2() {
	s := make([]int, 3, 6)
	s1 := SliceAppend(s)
	fmt.Printf("[exp2] SliceAppend s: %v, cap: %d \n", s, cap(s))      //[SliceParamExp2] SliceAppend s: [0 0 0], cap: 6
	fmt.Printf("[exp2] SliceAppend s1: %v, s1 cap: %d\n", s1, cap(s1)) //[SliceParamExp2] SliceAppend s1: [0 0 0 100], s1 cap: 6
}

func SliceAdd(s []int) {
	// 改变切片底层数组中的值，会志杰影响函数外层的变量（ps：参数s，在函数内改变后影响到外层s变量了，那函数参数是引用嘛？）
	for i := 0; i < len(s); i++ {
		s[i] += 1
	}
}

func SliceAppend(s []int) []int {
	// 这里s虽然改变了，但不会影响函数外层的s
	s = append(s, 100)
	return s
}
