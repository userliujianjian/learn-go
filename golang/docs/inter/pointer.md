### 值接收者和指针接受者的区别  
在创建一个结构体`struct`声明一个方法时有没有考虑或纠结过这两种写法`func (c *xxx)xxx... ， func(c xxx) xxxx...`用哪种？

让我们看下面的示例，一起看看有什么不同：  
```go

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
```
#### 分析
让我门看看 `PointerCoder` 方法中关于`c`变量两种不同的声明方式，思考一下为什么第一种值类型编译不通过呢？
我们分析一下：报错是说`Gopher` 没有实现`coder`,很明显了吧，因为 `Gopher`类型没有实现 `debug()`方法; 表面上看， `*Gopher` 类型也没有实现 `code()`方法，但是因为 `Gopher`类型实现了`code()`方法，所以让`*Gopher`类型自动拥有了`code()`方法。  
- 简单的解释：
	接收者是指针类型的方法，很可能在方法中会对接收者的属性进行更改操作，从而影响接收者；  
	而对接收者是值类型方法，在方法中不会对接收者本身产生影响  
** 所以，当实现了一个接收者是值类型的方法，就可以自动生成一个接收者是指针对应的方法，因为两者都不会影响接收者。但是，当实现了一个接收者是指针类型的方法，如果此时自动生成一个接收者是值类型的方法，原本期望对接收者改变（通过指针实现），现在无法实现，因为值类型会产生拷贝，不会真正影响调用者**  
> **如果实现了接收者是值类型的方法，会隐含地也实现接收者是指针类型的方法**  

#### 值类型和指针类型区别 
Q: 声明类型之前需要考虑一个问题，这个类型的本质是什么？如果给这个类型增加或删除某个值，是要创建一个新值，还是要更改当前的值？ 
- 值接收者：如果是要创建一个新值，该类型的方法就是用值接收者。  
- 指针接收者：如果是要修改当前值，就是用指针接收者
这个答案也会影响程序内部传递这个类型的方式：是按照值传递，还是按指针做传递。 保持传递的一致性很重要  
> 这个背后的原则是，**不要只关注某个方法是如何处理这个值，而是要关注这个值的本质是什么（增删改查）**

#### 使用指针作为方法的接收者理由：
- 方法能够修改接收者指向的值（改）  
- 避免每次调用方法时复制该值，在值的类型为大引用结构体时，这样做会更佳高效  

