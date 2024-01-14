package main

import (
	"sync"
	"time"
)

func unbufferedChannel() {
	c := make(chan string)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c <- `foo`
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		println(`Message: ` + <-c)
	}()

	wg.Wait()
}

func bufferedChannel() {
	c := make(chan string, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c <- `foo`
		c <- `bar`
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		println(`buffered Message: ` + <-c)
		println(`buffered Message: ` + <-c)
	}()

	wg.Wait()
}

//func main() {
//	unbufferedChannel()
//	bufferedChannel()
//
//}
