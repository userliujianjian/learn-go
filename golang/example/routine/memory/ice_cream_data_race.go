package main

import (
	"fmt"
	"runtime"
)

type IceCreamMaker interface {
	// Helo greets a customer
	Hello()
}

type Ben struct {
	name string
}

func (b *Ben) Hello() {
	fmt.Printf("Ben says, \"Hello my name is %s \" \n", b.name)
}

type Jerry struct {
	nickname string
}

func (j *Jerry) Hello() {
	fmt.Printf("Jerry says, \"Hello my name is %s \" \n", j.nickname)
}

func loop() {
	runtime.GOMAXPROCS(2)

	var ben = &Ben{"Ben"}
	var jerry = &Jerry{"Jerry"}
	var maker IceCreamMaker = ben

	var loop0, loop1 func()

	loop0 = func() {
		maker = ben
		go loop1()
	}

	loop1 = func() {
		maker = jerry
		go loop0()
	}

	go loop0()
	for i := 0; i < 200; i++ {
		maker.Hello()
	}

}

//func main() {
//	loop()
//}
