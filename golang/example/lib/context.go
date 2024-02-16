package lib

import (
	"fmt"
	"net/http"
	"time"
)

func gen() <-chan int {
	ch := make(chan int)
	go func() {
		var n int
		for {
			ch <- n
			n++
			time.Sleep(time.Second)
		}
	}()
	return ch
}

// 如果Goroutine
func main(w http.ResponseWriter, r *http.Request) {
	for n := range gen() {
		fmt.Println(n)
		if n == 5 {
			break
		}
	}

	// .... 业务代码
}

func test() {

}
