package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkMutexMultipleReaders(b *testing.B) {
	var lastValue uint64
	var lock sync.RWMutex

	cfg := Config{
		a: []int{0, 0, 0, 0, 0, 0},
	}

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				lock.RLock()
				atomic.SwapUint64(&lastValue, uint64(cfg.a[0]))
				lock.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
