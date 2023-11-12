package inter

func Init() {
	dog := Dog{}
	cat := Cat{}

	sayHello(dog)
	sayHello(cat)
	// 值传递
	pointerExp()
	usedTest()
}
