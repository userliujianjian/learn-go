package main

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

//func main() {
//	//example()
//	//example_1_solution()
//}

func example() {
	done := make(chan bool)
	m := make(map[string]string)
	m["name"] = "world"
	go func() {
		m["name"] = "data race"
		done <- true
	}()

	//<-done
	fmt.Println("Hello, ", m["name"])
	// data race 的原因是因为goroutine 跟主goroutine同时写map
	<-done // 无缓冲通道，当接收准备好之后，发送才开始执行。
}

func example1() {
	start := time.Now()
	var t *time.Timer
	t = time.AfterFunc(randomDuration(), func() {
		fmt.Println(time.Now().Sub(start))
		t.Reset(randomDuration())
	})
	time.Sleep(5 * time.Second)
}

func randomDuration() time.Duration {
	return time.Duration(rand.Int63n(1e9))
}

func example_1_solution() {
	start := time.Now()
	reset := make(chan bool)
	var t *time.Timer

	t = time.AfterFunc(randomDuration(), func() {
		fmt.Println(time.Now().Sub(start))
		reset <- true
	})

	for time.Since(start) < time.Second*5 {
		<-reset
		t.Reset(randomDuration())
	}
}

var blackHole [4096]byte // shared buffer

func ReadFrom(r io.Reader) (n int64, err error) {
	readSize := 0
	for {
		readSize, err = r.Read(blackHole[:])
		n += int64(readSize)
	}
	if err != nil {
		if err == io.EOF {
			return n, nil
		}
		return
	}
	return
}
