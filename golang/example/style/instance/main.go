package main

import (
	"fmt"
	"sync"
)

// 单例模式

type Once chan struct{}

func NewOnce() Once {
	o := make(Once, 1)
	o <- struct{}{}
	return o
}

func (o Once) Do(f func()) {
	_, ok := <-o
	if !ok {
		fmt.Println("pass...")
		return
	}

	f()

	close(o)
}

type Singleton struct {
	Data string
}

var (
	instance *Singleton
	once     sync.Once
)

func GetInstance(i int) *Singleton {
	//oc := NewOnce()
	//oc.Do(func() {
	once.Do(func() {
		instance = &Singleton{
			Data: fmt.Sprintf("hello， I am a singleton！ %d", i),
		}
	})
	return instance
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			singleton := GetInstance(num)
			fmt.Printf("Instance Data: %s \n", singleton.Data)
		}(i)
	}
	wg.Wait()
}
