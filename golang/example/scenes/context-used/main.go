package main

import (
	"context"
	"fmt"
)

type message struct {
	responseChan chan<- int
	parameter    string
	ctx          context.Context
}

func ProcessMessages(work <-chan message) {
	for job := range work {
		select {
		// If the context is finished, don't bother processing the message
		case <-job.ctx.Done():
			continue
		default:
		}

		hardToCalculate := len(job.parameter)
		select {
		case <-job.ctx.Done():
		case job.responseChan <- hardToCalculate:

		}
	}
}

func newRequest(ctx context.Context, input string, q chan<- message) {
	r := make(chan int)

	select {
	// If the context finishes before we can send msg onto q,
	// exit early
	case <-ctx.Done():
		fmt.Println("Context ended before q could see message")
		return
	case q <- message{
		responseChan: r,
		parameter:    input,
		// We are placing a context in a struct. this is ok since it
		// is only stored as a passed message and we want q to know
		// when it can  discard this message
		ctx: ctx,
	}:

	}
}

func main() {
	q := make(chan message)
	go ProcessMessages(q)
	ctx := context.Background()
	newRequest(ctx, "hi ", q)
	newRequest(ctx, "hello", q)
	close(q)

}
