package main

import "fmt"

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()

	return out
}
func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()

	return out
}

func main() {
	c := gen(2, 3)
	out := sq(c)

	fmt.Println(<-out) // output: 4
	fmt.Println(<-out) // output: 9

	for n := range sq(sq(gen(2, 3))) {
		fmt.Println(n)
	}
}
