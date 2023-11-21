package ch

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Result struct {
	record string
	err    error
}

func search(term string) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "some value", nil
}

func process(term string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := make(chan Result, 1)

	go func() {
		record, err := search(term)
		ch <- Result{record: record, err: err}
	}()

	select {
	case <-ctx.Done():
		return errors.New("search canceled")
	case result := <-ch:
		if result.err != nil {
			return result.err
		}
		fmt.Println("Received: ", result.record)
		return nil
	}
}
