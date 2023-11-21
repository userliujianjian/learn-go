package main

import "fmt"

// 内存逃逸分析
// go build -gcflags "-m -l"
func main() {
	fmt.Println("Called stackAnalysis: ", stackAnalysis())
}

func stackAnalysis() int {
	data := 55
	return data
}
