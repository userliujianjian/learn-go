package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

var Shutdown int32 = 0

func DebugSelect() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-sigChan:
			Shutdown = 1
			continue
		case <-func() chan struct{} {
			complete := make(chan struct{})
			go LaunchProcessor(complete)
			return complete
		}():
			return

		}
	}

}

func DebugSelect2() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	complete := make(chan struct{})
	go LaunchProcessor(complete)

	for {
		select {
		case <-sigChan:
			Shutdown = 1
			continue
		case <-complete:
			return

		}
	}
}

func debugSelectMain() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	complete := make(chan struct{})
	go LaunchProcessor(complete)

	for {
		select {
		case <-sigChan:
			atomic.StoreInt32(&Shutdown, 1)
			continue
		case <-complete:
			return

		}
	}
}

func LaunchProcessor(complete chan struct{}) {
	defer func() {
		close(complete)
	}()

	fmt.Printf("Start work \n")

	for count := 0; count < 5; count++ {
		fmt.Printf("Doing Work \n")
		time.Sleep(time.Second)
		if atomic.LoadInt32(&Shutdown) == 1 {
			fmt.Printf("Kill Early \n")
			return
		}
	}
}

func launchProcessor(complete chan struct{}) {
	defer func() {
		close(complete)
	}()

	fmt.Printf("Start work \n")

	for count := 0; count < 5; count++ {
		fmt.Printf("Doing Work \n")
		time.Sleep(time.Second)
		if Shutdown == 1 {
			fmt.Printf("Kill Early \n")
			return
		}
	}
}
