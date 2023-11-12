package inter

import "fmt"

type Animal interface {
	SayHello()
}

func sayHello(a Animal) {
	a.SayHello()
}

type Dog struct{}

func (d Dog) SayHello() {
	fmt.Println("wang wang wang....")
}

type Cat struct{}

func (c Cat) SayHello() {
	fmt.Println("miao miao miao....")
}
