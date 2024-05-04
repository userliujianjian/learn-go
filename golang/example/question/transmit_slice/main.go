package main

import "fmt"

// 切片作为函数参数传递
func updateSliceItem(data []string, index int, val string) {
	data[index] = val
}

func addSliceItem(data []string, val string) []string {
	data = append(data, val)
	return data
}

func main() {
	data := []string{"LJJ", "Bob", "Jon"}
	fmt.Println("step 1 data: ", data)
	updateSliceItem(data, 0, "zhangsan")
	fmt.Println("step 2 修改后的数据 data:", data)

	data = addSliceItem(data, "A")
	fmt.Printf("step 3 修改后的数据 data 值: %v, 长度: %d, 容量: %d \n", data, len(data), cap(data))
	_ = addSliceItem(data, "B")
	fmt.Printf("step 4 修改后的数据 data 值: %v, 长度: %d, 容量: %d \n", data, len(data), cap(data))
}
