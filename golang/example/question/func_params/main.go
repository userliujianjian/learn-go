package main

import "fmt"

func initMapData(data map[string]string) map[string]string {
	data["age"] = ""
	data["name"] = ""
	data["sex"] = ""
	data["school"] = ""
	return data
}

// 测试字典作为参数传递运行结果会如何
func main() {
	info := make(map[string]string)
	fmt.Println("step1 info Length: ", len(info))
	info = initMapData(info)
	fmt.Println("step2 info length", len(info))

	infoV2 := make(map[string]string)
	fmt.Println("step3 infoV2 Length: ", len(infoV2))
	initMapData(infoV2)
	fmt.Println("step4 infoV2 length", len(infoV2))
}
