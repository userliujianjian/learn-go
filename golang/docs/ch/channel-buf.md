## Goroutine泄漏

### 介绍
并发编程允许开发人员使用多个执行路径来解决问题，并且通过用于尝试提高性能。并发并不意味着这些多条路径并行执行；而是意味着这些多条路径是并行执行的。这意味着这些路径是无序的，而不是按照顺序执行的。从历史上看，这种类型的编程是通过使用标准库或第三方人员提供的库来实现的

在go中，Goroutines和通道等并发功能内置于语言和运行时中，以减少或消除对库的需求。这造成了一种错觉：用Go编写并发程序很容易。再决定使用并发时必须谨慎，因为如果使用不当，它会带来一些独特的副作用或陷阱。如果您不小心，这些陷阱可能会造成复杂性和令人讨厌的错误。

这篇文章中讨论的陷阱与Goroutine泄露有关

### 泄露Gotoutines
当涉及到内存管理时，Go会为您处理许多细节。Go编译器使用转译分析来决定值在内存中的位置。运行时通常使用垃圾收集器来跟踪和管理堆分配。尽管在应用程序中产生内存泄漏并非不可能，但可能性大大降低了。

一种常见的内存泄漏类型是Goroutines泄漏。 如果您启动了一个希望最终终止的Goroutines，但它从未停止，那么它就已经泄漏了。它在应用程序的生命周期内存在，并且为Goroutine分配的任何内存都无法释放。这是**永远不要在不知道Goroutine将如何停止的情况下启动它**的建议背后的部分原因。

为了说明基本Goroutine泄漏，请查看以下代码：
- 清单 1
```go
func leak() {
	ch := make(chan int)

	go func(){
		val := <-ch
		fmt.Println("We received a value: ", val)
	}()
}

```
清单1定义了一个名为leak的函数，该函数第二行创建一个通道，允许Goroutines传递证书数据。然后在第4行创建了Goroutine，第5行阻塞等待从通道接收值。当Goroutine等待时，leak函数返回。此时，程序的任何其他部分都无法通过该通道发送信号。这使得Goroutine在第6行被阻塞，无期限等待。第6行`fmt.Println`永远不会发生。

在此示例中，可以在代码审查期快速识别Goroutine泄漏。不幸的是，生产代码中的Goroutine泄漏通常更难发现。我无法展示Goroutine泄漏发生的所有可能方式，但这篇文章将详细介绍您可能遇到的一种Goroutine泄漏：

### 泄漏：被遗忘的发件人
> 对于这个泄漏示例，您将看到一个无限期阻塞的Goroutine，等待通道上发送值  
我们将看到程序根据某个搜索词查找一条记录，然后打印该记录。该程序是围绕一个名为Search函数构建的

- 清单2
```go
func search(term string) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "some value", nil
}
```
清单2中，第1行的函数 search 是一个模拟实现，用于模拟长时间运行的操作，例如数据库查询或web调用。在本示例中，编码为200ms  
该程序调用search函数，如清单3所示。 

- 清单3
```go
func process(term string) error{
	record, err := search(term)
	if err != nil {
		return err
	}

	fmt.Println("Received : ", record)
	return nil
}
```
在清单3的第一行中，定义了一个名为process，第二行将参数term单个参数传递给search返回记录和错误的函数。  
如果发生错误，第3行将返回错误给第用者。如果没有错误，则在第7行打印记录。

对于某些应用程序，顺序调用时产生的延迟search可能是不可接受的。  
假设无法使search函数运行的更快，process则可以将该函数更改为不消耗所有生产的总延迟成本

为此可以使用goroutine，如下面清单4所示。 不幸的是，第一次尝试是有问题的，因为它会造成潜在的goroutine泄漏。

- 清单4
```go
type Result struct {
	record string
	err    error
}

func search(term string) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "some value", nil
}

func process(term string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := make(chan Result)

	go func() {
		record, err := search(term)
		ch <- Result{record: record, err: err}
	}()

	select {
	case <-ctx.Done():
		return errors.New("search canceled")
	case result := <-ch:
		if result.err != nil {
			return result.err
		}
		fmt.Println("Received: ", result.record)
		return nil
	}
}
```
在清单4, 重写了process函数，创建Context将在100毫秒内取消函数。有关如何使用Context更多信息请参阅：[golang.org](https://go.dev/blog/context)

然后创建一个无缓冲通道ch，允许goroutines传递该result类型数据。接下来创建了一个Goroutine，调用search并尝试给ch无缓冲通道发送其返回值。

当Goroutine工作时，process函数执行select，有两种情况，都是通道接收操作。
有一个 ctx.Done()通道接收情况，context如果取消（100毫秒持续时间过去），则会执行此情况。如果执行这个case， 那么process会返回一个错误，表明已放弃等待search函数。  
或者第二个case，从ch通道接收并将值分配给名为result的变量。与之前的顺序实现一样，程序检查并处理错误。如果没有错误程序将打印 received并返回nil以提示成功。

process此重构设置了函数等待search完成的最长持续时间。然而这种实现也造成了潜在的goroutine泄。想想这段代码中goroutine是做什么的； 
在`ch <- result{record, err}` 它在给通道发送，此通道上会发生阻塞执行，知道另一个goroutine准备好接收该值。  
在超时的情况下，接收者停止等待从goroutine接收并继续。这将导致**Goroutine永远阻塞**，等待接收者出现，但这种情况永远不会发生。这就是goroutine泄漏的时候


### 修复： 腾出一些空间
解决此泄漏最简单的方法就是将通道从无缓冲通道更改为容量为1的缓冲通道。

- 清单5
```go
ch := make(chan Result, 1)
```

现在接收超市的情况下，在接收者继续前进后，搜索goroutine将通过值放入result通道中来完成其发送，然后返回。该goroutine的内存以及通道的内存最终都会被回收。一切都会自然而然的解决。

在[渠道行为一书中](https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html) 给了我们几个渠道行为的好例子，并提供了其相关使用的哲学。


### 结论
Go使启动Goroutine变得简单，但我们有责任明智地使用它们。在这篇文章中，我战士了一个如何错误使用goroutine的示例。 还有很多方法可以创建goroutine泄漏以及使用并发时可能遇到的其他陷阱。在以后的文章中，我将提供更多goroutine泄漏和其他并发陷阱示例。  
现在我将给你这个建议： 任何时候你启动一个goroutine时你必须问自己：
	- 什么时候回终止？
	- 什么可以阻止它终止？
**并发是一个有用的工具，但必须谨慎使用**
### 扩展
- 写一段会产生死锁的代码（提示：同一个goroutine中多次加锁、使用channel一个传入，另一个等待输出，select等待）
```go
// 根据通道的
func lock(){
	ch := make(chan int)

	go func(){
		ch <- 1
	}()

	go func(){
		<-ch
	}()

	select{}
}


func lock2(){
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(1)
	go func(){
		defer wg.Done()

		// 在goroutine中先锁住互斥锁
		mu.Lock()
		defer mu.Unlock()

		// 尝试再次锁住互斥锁，这回导致死锁
		mu.Lock()

		fmt.Println("这一行无法被打印")


	}()

	wg.Wait()


}
```

- 写出你知道的并发方式
```go

```