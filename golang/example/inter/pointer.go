package inter

import "fmt"

type Person struct {
	age int
}

func (c Person) howOld() int {
	return c.age
}

func (c *Person) growUp() {
	c.age += 1
}

func pointerExp() {
	// 值类型
	p := Person{age: 18}
	//值类型 调用接收者也是值类型的方法
	fmt.Println(p.howOld())

	// 值类型 调用接收者是真真类型的方法
	p.growUp()
	fmt.Println(p.howOld())

	// stefno 指针类型
	stefno := &Person{age: 100}
	// 指针类型 调用接收者是值类型的方法
	fmt.Println(stefno.howOld())

	//指针类型 调用接收者也是指针类型的方法
	stefno.growUp()
	fmt.Println(stefno.howOld())

}

type coder interface {
	code()
	debug()
}

type Gopher struct {
	language string
}

func (p Gopher) code() {
	fmt.Printf("I am codeing %s language \n", p.language)
}

func (p *Gopher) debug() {
	fmt.Printf("I am debuging %s language \n", p.language)
}

func PointerCoder() {
	// 变量c的声明方式，第一种声明编译不通过
	//var c coder = Gopher{language: "Go"}
	var c coder = &Gopher{"Go"}
	c.code()
	c.debug()
}

type userInterface interface {
	get() string
	set(v string)
}

func set(u userInterface, v string) {
	u.set(v)
}

func get(u userInterface) string {
	return u.get()
}

type UserWeb struct {
	name string
}

func (w *UserWeb) get() string {
	return w.name
}

func (w *UserWeb) set(v string) {
	w.name = v
}

type UserMobile struct {
	name string
}

func (m UserMobile) get() string {
	return m.name
}

func (m UserMobile) set(v string) {
	m.name = v
}

func usedTest() {
	var u userInterface = &UserWeb{name: "ljj"}

	set(u, "silent")
	name1 := get(u)
	fmt.Printf("[usedTest] name1: %v \n", name1)

	var u2 userInterface = UserMobile{"u2"}
	set(u2, "u2")
	fmt.Printf("[usedTest] u2: %v \n", u2)

	// 检测u2设置的值是否保留
	var u3 userInterface = UserMobile{}
	fmt.Println(get(u3))
	fmt.Printf("[usedTest] u3: %v \n", u3)

}
