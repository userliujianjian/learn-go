package main

import (
	"context"
	"io/ioutil"
	"runtime/trace"
	"sync"
)

func userTrace() {
	ctx, task := trace.NewTask(context.Background(), "main start")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		r := trace.StartRegion(ctx, "reading file")
		defer r.End()

		ioutil.ReadFile("n1.txt")
	}()

	go func() {
		defer wg.Done()
		r := trace.StartRegion(ctx, "writing file")
		defer r.End()

		ioutil.WriteFile(`n2.txt`, []byte(`42`), 0644)
	}()

	wg.Wait()

	defer task.End()
}
