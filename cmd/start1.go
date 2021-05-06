package main

import (
	"fmt"
	"log"
	"net/http"
)

func main1() {
	// 注册路由
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello GopherCon SG")
	})
	// 后台开启 8080 监听
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// 空的select语句永远阻塞
	select {}
}
