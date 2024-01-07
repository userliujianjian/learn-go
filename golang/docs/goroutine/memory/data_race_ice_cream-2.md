## 冰激淋制造者引起的数据竞争二

### 介绍
戴夫~切尼发表了一篇名为[《冰激淋制造商和数据竞赛》](./data_race_ice_cream.md)的文章。 这篇文章展示了一个有趣的数据竞争示例，当使用接口类型话变量进行方法调用时，可能会发生这种情况。如果您还没有阅读这篇文章，请阅读。一旦你读了这篇文章，你就会发现问题在于，接口值时使用这两个字的标题在内部实现的，而Go内存模型状态只写一个字是原子的。  

帖子中的程序显示了一个竞争条件，该条件允许两个goroutine同时对接口值进行读写操作。不同步此读取和写入允许读取观察到对接口值的部分写入。这允许Ben类型的实现针对Jerry类型的值进行操作，反之亦然。  

在Dave的例子中，Ben和Jerry结构的布局在内存中是相同的，因此它们在某种意义上时兼容的。Dave认为，如果它们由不同的记忆表征，就会发生混乱。这是因为Hello方法的每个实现都假定代码针对接收器类型的值进行操作。当这个错误付出水面时，情况就不再如此了。为了让你直观地了解这个建议，我们将两种不同的方式更改Jerry类型的生命。这两项更改将使您更好地理解接口类型和内存的互通。  

第一次代码更改：  
让我们回顾一下代码，看看第一组更改。

```go
package main

import (
	"fmt"
	"runtime"
)

type IceCreamMaker2 interface {
	// Great a customer.
	Hello()
}

type Ben2 struct {
	name string
}

func (b *Ben2) Hello() {
	if b.name != "Ben" {
		fmt.Printf("Ben says, \"Hello my name is %s \" \n", b.name)
	}
}

type Jerry2 struct {
	field1 *[5]byte
	field2 int
}

func (j *Jerry2) Hello() {
	name := string((*j.field1)[:])

	if name != "Jerry" {
		fmt.Printf("Jerry says, \"Hello my name is %s \" \n", name)
	}
}

func main() {
	runtime.GOMAXPROCS(2)
	var ben = &Ben2{"Ben"}
	var jerry = &Jerry2{&[5]byte{'J', 'e', 'r', 'r', 'y'}, 5}

	var maker IceCreamMaker2 = ben
	var loop0, loop1 func()

	loop0 = func() {
		maker = ben
		go loop1()
	}

	loop1 = func() {
		maker = jerry
		go loop0()
	}

	go loop0()

	for i := 0; i < 1000; i++ {
		maker.Hello()
	}
}

```  
在Ben2声明的Hello方法实现中，我将代码更改为尽在名称不是Ben的时候现实消息。这是一个简单的更改，因此我们不必通过结果来寻找错误合适出现。  
然后完全改变了Jerry2类型的声明。声明现在时字符串的手动表示形式。Go中的字符串由一个包含两个单词的标头组成。第一个单词是指向字节数组的指针，第二个单词是字符串的长度。这类似于切片，但容量的标题中没有第三个单词。Ben2和Jerry2结构的声明表示相同的内存布局，尽管生命方式非常不同。  

最后在创建并初始化Jerry2类型的变量，设置字节和长度。然后代码其余部分保持原样。  

当我们运行这个新版本时，输出根本不会改变：

```text
Jerry says, "Hello my name is Ben"
Ben says, "Hello my name is Jerry"
Ben says, "Hello my name is Jerry"
Jerry says, "Hello my name is Ben"
Ben says, "Hello my name is Jerry"
```
尽管Ben和Jerry类型的声明不同，**但内存布局是相同的**，并且程序按设计运行


```go
type Ben struct {
   name string
}

type Jerry struct {
   field1 *[5]byte
   field2 int
}

fmt.Printf("Ben says, \"Hello my name is %s\"\n", b.name)

name := string((*j.field1)[:])
fmt.Printf("Jerry says, \"Hello my name is %s\"\n", name)
```  
在对Hello方法的Ben类型实现进行Printf函数调用时，代码认为b指针指向Ben类型的值，而实际并非如此。但是，由于Ben和Jerry类型之间的内存布局相同，因此对Printf函数的调用仍然有效。 Jerry类型的Hello方法的实现也是如此。field1和field2的值等同于声明字符串字段，因此一切正常。  

```go
package main

import (
	"fmt"
)

type IceCreamMaker2 interface {
	// Great a customer.
	Hello()
}

type Ben2 struct {
	name string
}

func (b *Ben2) Hello() {
	if b.name != "Ben" {
		fmt.Printf("Ben says, \"Hello my name is %s \" \n", b.name)
	}
}

type Jerry2 struct {
	field2 int
	field1 *[5]byte
}

func (j *Jerry2) Hello() {
	name := string((*j.field1)[:])

	if name != "Jerry" {
		fmt.Printf("Jerry says, \"Hello my name is %s \" \n", name)
	}
}

func main() {
	//runtime.GOMAXPROCS(2)
	var ben = &Ben2{"Ben"}
	//var jerry = &Jerry2{&[5]byte{'J', 'e', 'r', 'r', 'y'}, 5}
	var jerry = &Jerry2{5, &[5]byte{'J', 'e', 'r', 'r', 'y'}}

	var maker IceCreamMaker2 = ben
	var loop0, loop1 func()

	loop0 = func() {
		maker = ben
		go loop1()
	}

	loop1 = func() {
		maker = jerry
		go loop0()
	}

	go loop0()

	for i := 0; i < 1000; i++ {
		maker.Hello()
	}
}
```  

现在，Jerry2生命切换了两个字段成员的顺序。整数值现在位于字节数组指针之前。当我们运行这个版本的程序时，我们会得到一个堆栈跟踪：  
```bash
Ben: 0x20817a170 Jerry: 0x20817a180
Ben：0x20817a170 Jerry：0x20817a180

01 panic: runtime error: invalid memory address or nil pointer dereference
02 [signal 0xb code=0x1 addr=0x5 pc=0x294f6]
03
04 goroutine 16 [running]:
05 runtime.panic(0xb90e0, 0x144144)
06    /Users/bill/go/src/pkg/runtime/panic.c:279 +0xf5
07 fmt.(*fmt).padString(0x2081b42d0, 0x5, 0x20817a190)
08    /Users/bill/go/src/pkg/fmt/format.go:130 +0x390
09 fmt.(*fmt).fmt_s(0x2081b42d0, 0x5, 0x20817a190)
10    /Users/bill/go/src/pkg/fmt/format.go:285 +0x67
11 fmt.(*pp).fmtString(0x2081b4270, 0x5, 0x20817a190, 0x73)
12    /Users/bill/go/src/pkg/fmt/print.go:511 +0xe0
13 fmt.(*pp).printArg(0x2081b4270, 0x97760, 0x20817a210, 0x73, 0x0, 0x0)
14    /Users/bill/go/src/pkg/fmt/print.go:780 +0xbb8
15 fmt.(*pp).doPrintf(0x2081b4270, 0xddfd0, 0x20, 0x220832de40, 0x1, 0x1)
16    /Users/bill/go/src/pkg/fmt/print.go:1159 +0x1ecc
17 fmt.Fprintf(0x220818c340, 0x2081c2008, 0xddfd0, 0x20, 0x220832de40, 0x1, 0x1, 0x10, 0x0, 0x0)
18     /Users/bill/go/src/pkg/fmt/print.go:188 +0x7f
19 fmt.Printf(0xddfd0, 0x20, 0x220832de40, 0x1, 0x1, 0x5, 0x0, 0x0)
20    /Users/bill/go/src/pkg/fmt/print.go:197 +0xa2
21 main.(*Ben).Hello(0x20817a180)
22    /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:16 +0x118
23 main.main()
24    /Users/bill/Spaces/Go/Projects/src/github.com/goinaction/code/temp/main.go:54 +0x2c3
```  
如果我们查看堆栈跟踪信息，我们将看到Hello的方法调用如何使用Ben2类型的实现，但传递了Jerry类型的值的地址。在对战跟踪之前，我先试了每个值的地址以明确这一点。如果我们再看一遍Ben2和Jerry2类型的声明，我们可以看到它们如何不再兼容：  

```go
type Ben struct {
   name string
}

type Jerry struct {
   field2 int
   field1 *[5]byte
}
```  
由于Jerry类型的这个新声明现在以整数值开头，因此它与字符串类型不兼容。这一次，当代码尝试打印b.name值时，程序对战将进行跟踪。  

### 结论：
最后，正在运行的程序在没有编译器提供任何保护措施的情况下操纵内存，CPU将按照指示解释该内存。在崩溃示例中，由于数据争用错误，代码要求CPU将整数值解释为字符串，程序崩溃。  所以我同意戴夫的观点，没有安全的数据竞赛。**程序要么没有数据争用，要么其操作为定义**。


#### 参考文章：
[原文](https://www.ardanlabs.com/blog/2014/06/ice-cream-makers-and-data-races-part-ii.html)