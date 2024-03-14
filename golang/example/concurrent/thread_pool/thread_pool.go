package main

import (
	"bufio"
	"fmt"
	"github.com/goinggo/workpool"
	"os"
	"runtime"
	"strconv"
	"time"
)

type MyWork struct {
	Name      string
	BirthYear int
	WP        *workpool.WorkPool
}

func (mw *MyWork) DoWork(workRoutine int) {
	fmt.Printf("%s: %d \n", mw.Name, mw.BirthYear)
	fmt.Printf("Q: %d, R: %d \n", mw.WP.QueuedWork(), mw.WP.ActiveRoutines())

	// Simulate some delay
	time.Sleep(100 * time.Millisecond)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork{
				Name:      "A" + strconv.Itoa(i),
				BirthYear: i,
				WP:        workPool,
			}

			if err := workPool.PostWork("routine", &work); err != nil {
				fmt.Printf("ERROR: %s \n", err)
				time.Sleep(100 * time.Millisecond)
			}

			if shutdown {
				return
			}
		}
	}()

	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true

	fmt.Println("Shutting Down \n")
	workPool.Shutdown("name_routine")

}
