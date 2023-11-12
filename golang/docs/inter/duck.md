### Go语言与鸭子类型的关系
先来看维基百科里面的定义：  
> if it looks like a duck, swims like a duck, and quacks like a duck, then it probably is a duck.  
翻译过来就是：如果某个东西**长得像鸭子，像鸭子一样游泳，像鸭子一样嘎嘎叫**，那它就可以被看成是一只鸭子

`Duck Typing`, 鸭子类型，是动态编程语言的一种对象推断策略，它更关注对象能如何被使用，而不是对象的类型本身。  
Go作为一门静态语言，它通过借口的方式完美支持鸭子类型。  
例如，在动态语言Python中，定义一个这样的函数：
```python{2,1-5}
def hello_world(coder):
	coder.say_hello()

```
当调用次函数的时候，可以传入人意类型，只要它实现了 `say_hello` 函数就可以。如果没有实现，运行过程中就会报错。  
而在静态语言如Java，C++中，不需要显式声明实现了某个接口，之后才能用在任何需要这个接口的地方。如果你在程序中调用`hello_world`函数，却传入了一个根本没有实现`say_hello`的类型，那在**变异阶段就不会通过。这也是静态语言比动态语言更安全的原因**。  
动态语言和静态语言的差别在此就有体现，静态语言在编译期间就能发现类型不匹配的错误，不想动态语言，必须要运行到那一行代码才会报错。  
当然静态语言要求程序员在编码阶段就要按照规定编写程序，为每个变量规定数据类型，这在某种程度上，加大了工作量，也加长了代码

Go作为一门现代静态语言，是有后发优势的。它引入动态语言的遍历，同时又会进行静态语言的类型检查，写起来是非常Happy的，Go采用了这种的做法: **不要求类型显式地声明实现了某个接口，只要实现了相关方法即可，编译器就能检测到**  
先定义一个借口和使用此借口作为参数的函数:   
```go 
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

```

在main函数中，调用`sayHello()`函数时，传入 `dog, cat`对象，它们并没有显式地声明实现了Animal类型，只是实现了接口所规定的 SayHello() 函数。**实际上，编译器在调用 sayHello()函数时，会隐式地将`dog,cat`对象转换成 Animal类型，这也是静态语言的类型检查功能**

顺带再提一下动态语言的特点：
> Go作为一种静态语言，通过接口实现了`鸭子类型`，实际上Go的编译器在其工作中做了隐匿传递的转换工作

#### 参考资料：
[GO程序员面试笔记宝典](https://golang.design/go-questions/interface/duck-typing/)
