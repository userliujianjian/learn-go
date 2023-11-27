package lib

import (
	"fmt"
	"net/http"
)

// 清单1
func main() {
	hello := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "hello world")
	})

	http.Handle("/", hello)
	http.ListenAndServe(":8080", nil)
}
