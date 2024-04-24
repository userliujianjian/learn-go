package main

import "fmt"

// 函数的参数通过值传递-map

func initDataInfo(data map[string]string) {
	data["age"] = ""
	data["name"] = ""
	data["sex"] = ""
	data["school"] = ""
	return
}

// 测试字典作为参数传递运行结果会如何
func main() {
	info := make(map[string]string)
	fmt.Println("step1 info Length: ", len(info))
	initDataInfo(info)
	fmt.Println("step2 info length", len(info))
}
