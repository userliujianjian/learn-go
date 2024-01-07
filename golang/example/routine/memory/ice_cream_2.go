package main

import (
	"fmt"
)

type IceCreamMaker2 interface {
	// Great a customer.
	Hello()
}

type Ben2 struct {
	name string
}

func (b *Ben2) Hello() {
	if b.name != "Ben" {
		fmt.Printf("Ben says, \"Hello my name is %s \" \n", b.name)
	}
}

type Jerry2 struct {
	field2 int
	field1 *[5]byte
}

func (j *Jerry2) Hello() {
	name := string((*j.field1)[:])

	if name != "Jerry" {
		fmt.Printf("Jerry says, \"Hello my name is %s \" \n", name)
	}
}

func main() {
	//runtime.GOMAXPROCS(2)
	var ben = &Ben2{"Ben"}
	//var jerry = &Jerry2{&[5]byte{'J', 'e', 'r', 'r', 'y'}, 5}
	var jerry = &Jerry2{5, &[5]byte{'J', 'e', 'r', 'r', 'y'}}

	var maker IceCreamMaker2 = ben
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

	for i := 0; i < 1000; i++ {
		maker.Hello()
	}
}
