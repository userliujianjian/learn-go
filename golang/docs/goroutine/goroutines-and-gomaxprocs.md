## 并发、Goroutines和GoMAXPROCS

### 简介
要了解Go如何使便携并发程序变得更容易且不易出错，我们首先需要了解**什么是并发程序以及此类程序回导致的问题**，这篇文章将重点介绍什么是并发程序、Goroutine扮演的角色以及GOMAXPROCS环境变量和运行时函数如何影响Go运行和我们编写的程序的行为

#### 进程和线程
当我们运行一个应用程序时，比如我用来写这篇文章的浏览器，操作系统会为该应用程序创建一个进程。该进程的工作就像一个容器，容纳应用程序在运行时使用和维护的所有资源。这些资源包括内存地址空间、文件句柄、设备和线程等。  

线程是由操作系统调度等执行路径，用于执行我们在处理器中编写的代码。进程从一个线程（主线程）开始，当该线程终止时，进程也终止。这是因为主线程是应用程序的起源。然后主线程可以一次启动更多线程，并在这些线程上可以启动多个线程。  

操作系统安排线程在可用处理器上运行，无论该线程数语哪个进程。每个操作主系统都有自己的算法来做出这些决定，我们最好编写不特定于一种算法或另一种算法的并发程序。另外这些算法会锁着操作系统每个版本而变化，因此这是一个危险的游戏

#### Goroutine和并行性  
Go中的任何函数或方法都可以创建为Goroutine。我们可以认为main函数作为goroutine执行，但是go运行时不会启动该goroutine。Goroutines被认为是轻量级的，因为他们是用很少的内存和资源，而且他们的厨师堆栈很小。在go1.2之前，堆栈大小从4K开始，现在从1.4开始，堆栈大小从8k开始。该堆栈能够根据需要增长。

操作系统调度线程在可用处理器上运行，Go运行时将goroutine调度为在[逻辑处理器内运行](https://www.ardanlabs.com/blog/2015/02/scheduler-tracing-in-go.html)绑定到单个操作系统线程的逻辑处理器中运行。默认情况下，go运行时分配一个逻辑处理器来执行为我们的程序创建所有goroutine。即使使用这个单一的逻辑处理器和操作系统线程，也可以安排数十万个goroutine以惊人的效率和性能同时运行。不建议添加多个逻辑处理器，但如果您想并行运行goroutine，Go提供了GOMAXPROCS环境变量或运行时函数添加更多逻辑处理器的功能。  

并发不是并行。并行是指两个或多个线程同时针对不同处理器执行代码。在运行时配使用多个逻辑处理器，则调度程序将在这些逻辑处理器之间分配goroutine，这将导致goroutine在不同的操作系统线程上运行。但是要获得真正的并行性，您需要在具有多个物理处理器的计算机上运行程序。否则，即使GO运行时使用多个逻辑处理器，goroutines也将针对单个物理处理器并发运行。  

#### **并发示例**
让我们构建一个小程序，显示 Go 并发运行 goroutine。在此示例中，我们使用一个逻辑处理器运行代码
```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(1)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Starting Go Routines")
	go func() {
		defer wg.Done()

		time.Sleep(time.Microsecond)
		for char := 'a'; char < 'a'+26; char++ {
			fmt.Printf("%c ", char)
		}

	}()

	go func() {
		defer wg.Done()

		for number := 1; number < 27; number++ {
			fmt.Printf("%d ", number)
		}

	}()

	fmt.Println("Waiting TO Finish")
	wg.Wait()

	fmt.Println("\n Terminating Program")
}



```
该程序通过使用关键字go并声明两个匿名函数来启动两个goroutine。第一个goroutine使用小写字母显示英文字母，第二个goroutine现实数字1到266.当我们运行程序时，得到以下输出：
```text
Starting Go Routines
Waiting TO Finish
1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 a b c d e f g h i j k l m n o p q r s t u v w x y z 
Terminating Program

```

当我们查看输出时，我们可以看到代码是并发运行的。启动两个goroutine后，主goroutine将等待goroutines完成。我们需要这样做，是因为一旦主goroutine终止，程序就会终止。使用waitgroup是goroutines在完成后进行通信的最好办法。  

我们可以看到，一个goroutine完成显示所有26个字母，然后第二个goroutine轮到显示所有26个数字。因为第一个goroutine完成其工作时间不到1微妙，所以我们不会看到调度器在第二个goroutine完成工作之前中断它。我们可以给调度器一个理由，通过在第二个goroutine中加入睡眠来交换goroutines：
```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main2() {
	runtime.GOMAXPROCS(1)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Starting Go Routines")
	go func() {
		defer wg.Done()

		for char := 'a'; char < 'a'+26; char++ {
			fmt.Printf("%c ", char)
		}

	}()

	go func() {
		defer wg.Done()

		time.Sleep(100 * time.Microsecond)
		for number := 1; number < 27; number++ {
			fmt.Printf("%d ", number)
		}

	}()

	fmt.Println("Waiting TO Finish")
	wg.Wait()

	fmt.Println("\n Terminating Program")
}
```
这一次，我们第二个goroutine开始立即添加一个睡眠。调用sleep会导致调度程序交换两个goroutines：
```text
Starting Go Routines
Waiting TO Finish
a b c d e f g h i j k l m n o p q r s t u v w x y z 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 
 Terminating Program
```
这一次首先显示字母，然后是数字。休眠导致调度程序停止运行第二个goroutine，让第一个goroutine执行它的工作


### **并行示例**
在我们过去两个示例中，goroutines是**并发**运行的，但不是**并行**的。让我们对代码进行更改，允许goroutines并行运行。我们就需要做的就是在调度程序中添加第二个逻辑处理以使用两个线程：
```go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main3() {
	runtime.GOMAXPROCS(2)

	var wg sync.WaitGroup
	wg.Add(2)

	fmt.Println("Starting Go Routines")
	go func() {
		defer wg.Done()

		for char := 'a'; char < 'a'+26; char++ {
			fmt.Printf("%c ", char)
		}

	}()

	go func() {
		defer wg.Done()


		for number := 1; number < 27; number++ {
			fmt.Printf("%d ", number)
		}

	}()

	fmt.Println("Waiting TO Finish")
	wg.Wait()

	fmt.Println("\n Terminating Program")
}
```
这一次，我们第二个goroutine开始立即添加一个睡眠。调用sleep会导致调度程序交换两个goroutines：
```text
Starting Go Routines
Waiting TO Finish
a b c d e f g h i j k l m n o p q r s t u v w x y z 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 
 Terminating Program
```

每次我们运行程序时，我们都会得到不同的结果。调度程序在每次运行中的行为并不完全相同。我们可以看到goroutines确实是并行运行的。两个goroutine立即开始运行，您可以看到他们都在争标准输出以显示其结果。  

### 结论
仅仅因为我们可以添加多个逻辑处理器供调度程序使用，并不意味着我们都应该这样做。go团队以他们的方式默认设置为运行时是有原因的。尤其是仅使用单个逻辑处理器的默认值。要知道任意添加逻辑处理器和并行运行goroutines不一定会为程序提供良好的性能。始终对程序进行分析和基准测试，并确保在绝对需要时更改go运行时配置。  

在我们的应用程序中构建并发性的问题在于，最终我们的 goroutines 将尝试访问相同的资源，可能同时访问。对共享资源的读取和写入操作必须始终是原子操作。换句话说，读取和写入必须一次由一个 goroutine 发生，否则我们会在程序中创建竞争条件。  

通道是 Go 中我们编写安全优雅的并发程序的方式，它消除了竞争条件，让编写并发程序再次变得有趣。现在我们知道了 goroutine 是如何工作的、被调度的，并且可以并行运行，通道是我们需要学习的下一件事。

#### 参考文献：
https://www.ardanlabs.com/blog/2014/01/concurrency-goroutines-and-gomaxprocs.html



