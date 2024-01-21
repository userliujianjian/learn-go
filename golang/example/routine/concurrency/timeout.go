package main

import (
	"fmt"
	"time"
)

type Search struct {
}

func (s *Search) DoQuery(query string) string {
	time.Sleep(100 * time.Millisecond)
	return query
}

func main() {
	res := Test()
	fmt.Println(res)
	time.Sleep(time.Second)
}

func Test() string {
	ch := make(chan string)
	for i := 0; i < 10; i++ {
		search := Search{}
		go func(c Search) {
			select {
			case ch <- c.DoQuery("AAA"):
			default:
				fmt.Println("default time out")

			}
		}(search)
	}
	return <-ch
}
