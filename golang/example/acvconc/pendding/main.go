package main

import "fmt"

type Item struct {
	Title string
}

func main() {
	var pending []Item
	ch := make(chan Item)

	for {
		var first Item
		var updates chan Item
		if len(pending) > 0 {
			first = pending[0]
			updates = ch
		}

		select {
		case updates <- first:
			fmt.Println("for........ \n")
			pending = pending[1:]
		}

	}

}
