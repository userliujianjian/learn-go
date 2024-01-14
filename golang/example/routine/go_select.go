package main

import (
	"fmt"
	"time"
)

func orderingInSelect() {
	a := make(chan bool, 10)
	b := make(chan bool, 10)
	c := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		a <- true
		b <- true
		c <- true
	}

	for i := 0; i < 10; i++ {
		select {
		case <-a:
			fmt.Print(" < a")
		case <-b:
			fmt.Print(" < b")
		case <-c:
			fmt.Print(" < c")
		default:
			fmt.Print(" < default")

		}
	}
}

func orderingInSelect2() {
	a := make(chan bool, 10)
	b := make(chan bool, 10)
	c := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		a <- true
		b <- true
		c <- true
	}

	for i := 0; i < 10; i++ {
		select {
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-a:
			fmt.Print(" < a")
		case <-b:
			fmt.Print(" < b")
		case <-c:
			fmt.Print(" < c")
		default:
			fmt.Print(" < default")

		}
	}
}

func selectWaiting() {
	a := make(chan bool, 10)
	b := make(chan bool, 10)

	go func() {
		time.Sleep(time.Minute)
		for i := 0; i < 10; i++ {
			a <- true
			b <- true
		}
	}()

	for i := 0; i < 10; i++ {
		select {
		case <-a:
			fmt.Print("< a ")
		case <-b:
			fmt.Print("< b ")
		}
	}
}

func selectOneCase() {
	t := time.NewTicker(time.Minute)
	select {
	case <-t.C:
		fmt.Print("1 minute later...")
	default:
		fmt.Print("default branch")
	}
}
