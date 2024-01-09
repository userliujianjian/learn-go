package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Config struct {
	a []int
}

func UpdateConfig() {
	cfg := &Config{}

	// Write
	go func() {
		i := 0
		for {
			i++
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				fmt.Println(cfg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func updateConfigMutex() {
	cfg := &Config{}

	lock := sync.RWMutex{}

	// write
	go func() {
		var i int
		for {
			i++
			lock.Lock()
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
			lock.Unlock()
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				lock.RLock()
				fmt.Println(cfg)
				lock.RUnlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func updateByAtomic() {
	var v atomic.Value

	//writer
	go func() {
		var i int
		for {
			i++
			cfg := &Config{
				a: []int{i, i + 1, i + 2, i + 3, i + 4, i + 5},
			}
			v.Store(cfg)
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				cfg := v.Load()
				fmt.Println(cfg)
			}
			wg.Done()
		}()
	}

	wg.Wait()

}

func main() {
	//UpdateConfig()
	//updateConfigMutex()
	//updateByAtomic()
}
