package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

// never start a goroutine without knowing when it will stop(永远不要再不知道它合适停止的情况下启动一个goroutine)

func startPprof() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello, QCon!")
	})

	go http.ListenAndServe(":8001", http.DefaultServeMux)
	http.ListenAndServe(":8080", mux)

}

func serveApp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello, QCon!")
	})

	http.ListenAndServe(":8080", mux)

}

func serveDebug() {
	http.ListenAndServe(":8001", http.DefaultServeMux)
}

func startMain() {
	go serveDebug()
	serveApp()
}

func serveApp2() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "Hello QCon!")
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func serveDebug2() {
	if err := http.ListenAndServe(":8001", http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}

func startMain2() {
	go serveApp()
	go serveDebug()
	select {}
}

func serveApp3() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello QCon!")
	})
	return http.ListenAndServe(":8080", mux)
}

func serveDebug3() error {
	return http.ListenAndServe(":8001", http.DefaultServeMux)
}

func startMain3() {
	done := make(chan error, 2)
	go func() {
		done <- serveApp3()
	}()

	go func() {
		done <- serveDebug3()
	}()

	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			fmt.Printf("error: %v", err)
		}
	}
}

func serve(addr string, handler http.Handler, stop <-chan struct{}) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-stop
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()

}

func serveAppMaster(stop <-chan struct{}) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello QCon!")
	})

	return serve(":8080", mux, stop)
}

func serveDebugMaster(stop <-chan struct{}) error {
	return serve(":8001", http.DefaultServeMux, stop)
}

func mainMaster() {
	done := make(chan error, 2)

	stop := make(chan struct{})

	go func() {
		done <- serveAppMaster(stop)
	}()

	go func() {
		done <- serveDebugMaster(stop)
	}()

	var stopped bool
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			fmt.Printf("error: %v", err)
		}
		if !stopped {
			stopped = true
			close(stop)
		}
	}
}
