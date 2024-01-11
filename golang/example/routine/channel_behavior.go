package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func listing1() {
	ch := make(chan string)
	go func() {
		p := <-ch // Receive
		fmt.Println(p)
	}()

	ch <- "paper" // Send
}

// channel state
func listing2() {
	// A channel is in a nil state when it is declared to its zero value
	var ch chan string

	// A channel can be placed in a nil state by explicitly setting it to nil.
	ch = nil

	// ** open channel
	// A channel is in a open state when it's made using the built-in function make.
	ch = make(chan string)

	// ** closed channel

	// A channel is in a closed state when it's closed using the build-in function close.
	close(ch)
}

func waitForTask() {
	ch := make(chan string)

	go func() {
		p := <-ch
		// Employee performs work here.

		// Employee is done and free to go.

		fmt.Println(p)
	}()

	time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)

	ch <- "paper"
}

func waitForResult() {
	ch := make(chan string)

	go func() {
		time.Sleep(time.Duration(rand.Intn(500)) * time.Microsecond)

		ch <- "paper"

		// Employee is done and free to go.
	}()

	p := <-ch
	fmt.Println(p)
}

func fanOut() {
	emps := 20
	ch := make(chan string, emps)

	for e := 0; e < emps; e++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(200)) * time.Microsecond)
			ch <- "paper"
		}()
	}

	for emps > 0 {
		p := <-ch
		fmt.Println(p)
		emps--
	}
}

func selectDrop() {
	const cp = 5
	ch := make(chan string, cp)

	go func() {
		for p := range ch {
			fmt.Println("employee: received: ", p)
		}
	}()

	const work = 20
	for w := 0; w < work; w++ {
		select {
		case ch <- "paper":
			fmt.Println("manager: send ack")
		default:
			fmt.Println("manager : drop")

		}
	}

	close(ch)
}

func waitForTasks() {
	ch := make(chan string, 1)

	go func() {
		for p := range ch {
			fmt.Println("employee: working: ", p)
		}
	}()

	const work = 10

	for w := 0; w < work; w++ {
		ch <- "paper"
	}

	close(ch)
}

func withTimeout() {
	duration := 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ch := make(chan string, 1)

	go func() {
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		ch <- "paper"
	}()

	select {
	case p := <-ch:
		fmt.Println("work complete", p)

	case <-ctx.Done():
		fmt.Println("moving on")
	}
}
