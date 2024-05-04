package main

import (
	"fmt"
	"github.com/goinggo/jobpool"
	"time"
)

type WorkProvider1 struct {
	Name string
}

func (wp *WorkProvider1) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 1 : started: %s \n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 1 : Done: %s \n", wp.Name)
}

type WorkProvider2 struct {
	Name string
}

func (wp *WorkProvider2) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 2 : started: %s \n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 2 : Done: %s \n", wp.Name)
}

func main() {
	jobPool := jobpool.New(2, 1000)
	_ = jobPool.QueueJob("main", &WorkProvider1{"Normal Priority: 1"}, false)
	fmt.Printf("***************> QW: %d AR: %d \n", jobPool.QueuedJobs(), jobPool.ActiveRoutines())

	time.Sleep(1 * time.Second)
	jobPool.QueueJob("main", &WorkProvider1{"normal Priority: 2"}, false)
	jobPool.QueueJob("main", &WorkProvider1{"normal Priority: 3"}, false)

	jobPool.QueueJob("main", &WorkProvider2{"Normal Priority: 4"}, true)
	fmt.Printf("***************> QW: %d AR: %d \n", jobPool.QueuedJobs(), jobPool.ActiveRoutines())
	time.Sleep(15 * time.Second)

	jobPool.Shutdown("main")
}
