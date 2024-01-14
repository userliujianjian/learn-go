package main

import (
	"fmt"
	"time"
)

func swapState() {
	// create an unbuffered channel
	baton := make(chan int)

	// First runner to his mark
	go Runner(baton)

	// start the race
	baton <- 1

	// Give the runners time to race
	time.Sleep(500 * time.Millisecond)
}

func Runner(baton chan int) {
	var newRunner int

	// wait to receive the baton
	runner := <-baton

	// start running around the trace
	fmt.Printf("Runner %d Running With Baton \n", runner)

	// new runner to the line
	if runner != 4 {
		newRunner = runner + 1
		fmt.Printf("Runner %d To the line \n", newRunner)
		go Runner(baton)
	}

	// running around the track
	time.Sleep(100 * time.Millisecond)

	// is the race over
	if runner == 4 {
		fmt.Printf("Runner %d Finished, Race over \n", runner)
		return
	}

	// exchange th baton for the next runner
	fmt.Printf("Runner %d Exchange with runner %d \n", runner, newRunner)
	baton <- newRunner

}
