package main

import (
	"context"
	"errors"
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
	fmt.Println("111111")
	g.Go(ServerApp)
	fmt.Println("22222")
	g.Go(Signal)
	fmt.Println("333333")

	// 保持接收
	if err := g.Wait(); err != nil {
		fmt.Printf("context error %v\n", err)
	}
	fmt.Printf("main end ---------")

}

func Signal() error {
	// 接收 linux signal 信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChan
	log.Printf("Get Signal: %v\n", sig)
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//ctx.Done()
	//defer cancel()
	//log.Printf("ctx : %v\n", ctx)
	// TODO shutdown server 结束server
	return errors.New("asdf")
}

func ServerApp() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
		// 四舍五入 保留两位小数
		var (
			a = 100
			b = 100
		)

		fmt.Printf("res: %v\n", 1/a-b)
		dayStr := ""
		_, err := time.ParseInLocation("20060102150405", dayStr, time.Local)
		if err != nil {
			fmt.Printf("format err : %v\n", err)
			log.Fatal(err)
		}

	})
	return http.ListenAndServe("0.0.0.0:8080", mux)
}

func ServerDebug() error {
	return http.ListenAndServe("127.0.0.1:8000", http.DefaultServeMux)
}
