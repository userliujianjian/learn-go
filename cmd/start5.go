package main

import (
	"context"
	"fmt"
	"net/http"
)

func main5() {
	done := make(chan error, 2)
	stop := make(chan struct{})

	go func() {
		done <- ServerApp1(stop)
	}()

	go func() {
		done <- ServerDebug1(stop)
	}()
	var stopped bool
	for i := 0; i < cap(done); i++ {
		fmt.Printf("for done %v\n", done)
		if err := <-done; err != nil {
			fmt.Printf("error: %v\n", err)
		}
		if !stopped {
			stopped = true
			close(stop)
		}
	}
}

func Server(addr string, handler http.Handler, stop <-chan struct{}) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-stop // wait for stop signal
		s.Shutdown(context.Background())
	}()
	fmt.Printf("[server] address %v\n", addr)

	return s.ListenAndServe()

}

func ServerApp1(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
	})
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		var (
			s = 100
			b = 100
		)
		e := s + b
		h := (s - b) / 1
		fmt.Println(e, h)
	})
	addr := "0.0.0.0:8080"
	return Server(addr, mux, stop)
}

func ServerDebug1(stop <-chan struct{}) error {
	mux := http.DefaultServeMux
	addr := "127.0.0.1:8009"
	return Server(addr, mux, stop)
}
