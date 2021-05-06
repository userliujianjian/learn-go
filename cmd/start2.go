package main

import (
	"fmt"
	"net/http"
)

func main2() {
	// keep yourself busy or do the work yourself
	// 注册路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
	})
	go http.ListenAndServe("127.0.0.1:8000", http.DefaultServeMux)
	http.ListenAndServe("0.0.0.0:8080", mux)

}
