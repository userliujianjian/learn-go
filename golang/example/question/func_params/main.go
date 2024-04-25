package main

import (
	"fmt"
)

func initMapData(data map[string]string) {
	data["age"] = ""
	data["name"] = ""
	data["sex"] = ""
	data["school"] = ""
	return
}

func initTwoDimensional(data map[string]map[string]string) {
	data["v1"] = make(map[string]string)
	data["v1"]["age"] = ""
	data["v1"]["name"] = ""
	data["v1"]["sex"] = ""
	data["v1"]["school"] = ""
	data["v2"] = make(map[string]string)
	data["v2"]["age"] = ""
	data["v2"]["name"] = ""
	data["v2"]["sex"] = ""
	data["v2"]["school"] = ""
}

// 测试字典作为参数传递运行结果会如何
func main() {
	info := make(map[string]string)
	fmt.Println("step1 info Length: ", len(info))
	initMapData(info)
	fmt.Println("step2 info length", len(info))

	infoV2 := make(map[string]string)
	fmt.Println("step3 infoV2 Length: ", len(infoV2))
	initMapData(infoV2)
	fmt.Println("step4 infoV2 length", len(infoV2))

	//dimensional
	twoDimensional := make(map[string]map[string]string)
	fmt.Println("step5 twoDimensional Length: ", len(twoDimensional))
	initTwoDimensional(twoDimensional)
	fmt.Println("step6 twoDimensional Length: ", len(twoDimensional))
	for k, v := range twoDimensional {
		fmt.Println("step7 twoDimensional key: ", k, "length: ", len(v))
	}
}
