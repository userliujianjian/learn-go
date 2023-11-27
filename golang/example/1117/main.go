package main

import "fmt"

func main() {

	s := []int{1, 2, 3, 4}
	m := make(map[int]*int)
	for k, v := range s {
		m[k] = &v
	}
	fmt.Printf("m: %v \n", m)

	s1 := s[:1]
	s1 = append(s1, 5)
	for i := 0; i < len(s1); i++ {
		item := s1[i]
		fmt.Printf("s1 is: k:%d -> v: %d \n", i, item)
	}

	for i := 0; i < len(s); i++ {
		item := s[i]
		fmt.Printf("s is: k:%d -> v: %d \n", i, item)
	}

}
