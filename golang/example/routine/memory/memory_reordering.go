package main

import (
	"fmt"
	"runtime"
	"sync"
)

func reordering() {
	var wg sync.WaitGroup
	wg.Add(2)

	var x, y int

	go func() {
		defer wg.Done()
		x = 1
		fmt.Print("y:", y, " ")
	}()
	go func() {
		defer wg.Done()
		y = 1
		fmt.Print("x:", x, " ")
	}()

	wg.Wait()

}

func main() {
	runtime.GOMAXPROCS(2)
	for i := 0; i < 10000; i++ {
		reordering()
		fmt.Print("\n")
	}

}
