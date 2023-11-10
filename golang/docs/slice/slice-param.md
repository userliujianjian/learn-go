### 切片作为函数参数

> **值得注意的是，Go 语言的函数参数传递，只有值传递，没有引用传递** 。

复习下前面所介绍的，slice结构体包含了三个成员：len,cap, array. 分别表示长度，容量，底层数据指针，非原子操作。
那么你对切片足够熟悉了嘛？我们一起看看下面的代码片段：  

```go
// exp1 切片作为参数，改变副本底层数据测试
func exp1() {
	s := []int{1, 2, 3}
	SliceAdd(s)
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
```


我们来看看 `exp1` 返回的结果`[2 3 4]` 果真**变量s的底层数据被改变**，这里传递的是一个s的副本，在函数 SliceAdd中，s只是exp1函数中的一个拷贝。 在 SliceAdd函数内部，**对s的修改并不会改变外层exp1函数中的s变量** 。  
为什么不会改变外层呢，我们来看`exp2` 返回结果，SliceAppend 虽然改变了s，但它只是一个值传递，并不会影响到外层的s，因此第一个打印 `[0 0 0]`。  
而s1 是一个新的slice，它基于s得到的，因此打印结果[0 0 0 100], 虽然容量没有改变，但是底层数据指针已经发生改变。  


参考文章：
https://golang.design/go-questions/slice/as-func-param/

