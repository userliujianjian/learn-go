package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Never start a goroutine without knowing when it will stop

func main() {
	// use errGroup 创建一个启动任务
	g, _ := errgroup.WithContext(context.Background())
	g.Go(ServerDebug)
	g.Go(ServerApp)

	// 保持接收
	if err := g.Wait(); err != nil {
		fmt.Printf("context error %v\n", err)
	}
	fmt.Printf("main end ---------")

}

func Server(addr string, handler http.Handler) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		// 接收 linux signal 信号
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		sig := <-signalChan
		log.Printf("Get Signal: %v\n", sig)
		s.Shutdown(context.Background())
	}()
	fmt.Printf("[server] address %v\n", addr)

	return s.ListenAndServe()

}

func ServerApp() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
	})
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		dayStr := ""
		_, err := time.ParseInLocation("20060102150405", dayStr, time.Local)
		if err != nil {
			fmt.Printf("format err : %v\n", err)
			log.Fatal(err)
		}
	})
	addr := "0.0.0.0:8080"
	return Server(addr, mux)
}

func ServerDebug() error {
	mux := http.DefaultServeMux
	addr := "127.0.0.1:8009"
	return Server(addr, mux)
}
