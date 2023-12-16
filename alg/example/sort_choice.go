package main

import "fmt"

// 选择排序算法O(n^2)
// 选择排序核心思想就是 拿着每一个元素跟每个元素做对比，
// 如果找到符合的 就跟当前元素调换位置。因为对比过之后就不需要再次加入对比，减少了创建新切片的过程
func selectSort(arr []int) []int {
	n := len(arr)
	// 外部的for循环最大是 n -1 是为了内部循环时 拿着当前元素去跟下一个元素做对比，
	// 所以内部的for循环最大值时N，开始值时i + 1
	for i := 0; i < n-1; i++ {
		minIndex := i
		// 将当前元素跟任意元素对比，找到最小的元素
		for j := i + 1; j < n; j++ {
			if arr[j] < arr[minIndex] {
				minIndex = j
			}
		}
		// 交换两个元素, 这样在J遍历的时候就不需要考虑i之前的元素了
		arr[i], arr[minIndex] = arr[minIndex], arr[i]
	}
	return arr
}

// 递归算法，找出n的阶乘
func factorial(n int) int {
	if n == 0 {
		return 1
	}

	return n * factorial(n-1)
}

// 找出目标数据的下标
// 主旨：跟中间数对比缩小范围
func binarySearch(arr []int, target int) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			low = mid + 1
		} else if arr[mid] > target {
			high = mid - 1
		}
	}
	return -1
}

// 快速排序法O(n^2), 基线条件：切片为空或者
func quickSOrt(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}
	pivot := arr[0]
	var less, greater []int
	for i := 1; i < len(arr); i++ {
		item := arr[i]
		if item <= pivot {
			less = append(less, item)
		} else {
			greater = append(greater, item)
		}
	}

	l := make([]int, len(arr))
	l = quickSOrt(less)
	l = append(l, pivot)
	l = append(l, quickSOrt(greater)...)
	return l

}

func sortInit() {
	// 二分查找发
	list := []int{1, 2, 3, 4, 5}
	index := binarySearch(list, 2)
	fmt.Printf("二分查找法binarySearch， 查找2的下标，运行结果：%v \n", index)

	// 选择排序
	arr := []int{1, 2, 6, 3, 4, 5}
	newArr := selectSort(arr)
	fmt.Printf("选择排序算法selectSort结果：%v \n", newArr)
	// 递归
	res := factorial(5)
	fmt.Printf("递归算法factorial 5阶乘的结果：%v \n", res)

	arr2 := []int{1, 2, 6, 3, 4, 5}
	newArr2 := quickSOrt(arr2)
	fmt.Printf("快速排序法quickSOrt 结果：%v \n", newArr2)

}
