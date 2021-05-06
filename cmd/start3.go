package main

import (
	"fmt"
	"net/http"
)

/**
Keep yourself busy or do the work busy
*/

func ServerApp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
	})
	http.ListenAndServe("0.0.0.0:8080", mux)
}

func ServerDebug() {
	http.ListenAndServe("127.0.0.1:8000", http.DefaultServeMux)
}

func main3() {
	go ServerDebug()
	// 如果serverApp 挂掉 main函数将会释放， serverDebug 也会结束
	ServerApp()
}
