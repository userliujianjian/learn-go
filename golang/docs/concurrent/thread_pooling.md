### Go编程中的线程池

在Go中工作了一段时间后，我学会了如何使用无缓冲通道来构建goroutines池。我比这篇文章中实现的更好。话虽如此，这篇文章所描述的内容仍然有价值。 

#### 介绍  
在我的服务器开发世杰中，线程池一直是在Microsoft堆栈上构建可靠代码的关键。Microsoft在.Net中失败了，它为每个进程提供了一个具有数千个线程的单线程池，并认为他们可以在运行时管理并发性。可早以前，我就意识到这永远行不通。至少对于我正在开发的服务器来说不是。  

当我是用Win32API在C/C++中构建服务器时，我创建了一个抽象IOCP的类，为我提供了可以将工作发布到的线程池。这一直很有效，因为我可以定义池中的线程数和并发级别（在任何给定时间允许处于活动状态的线程数）。我为我所有的C#开发移植了这段代码。如果你想了解更多，我几年前写了一篇[文章](http://www.theukwebdesigncompany.com/articles/iocp-thread-pooling.php). 使用IOCP为我提供了所需的性能和灵活性。顺便说一句，.NET线程池在线面使用IOCP。    

线程池的想法相当简单。工作进入服务器并需要处理。其中大部分工作本质上是异步的，但并非必须如此。很多时候，工作是从插座或内部例行程序中完成的。线程池将工作排队，然后从池中分配一个线程来执行工作。 工作按收到的顺序进行处理。游泳池尾了高效执行工作提供了一个很好的模式。每次需要处理工作时生成一个新线程可能会给操作系统带来沉重的负载，并导致重大的性能问题。   


那么线程池性能时如何调整的呢？您需要确定每个线程池包含的线程数，以最快完成工作。当所有例程都忙于处理工作时，新工作将保持排队状态。你想要这个，因为在某些时候，有更多的例程处理工作会减慢速度。这可能是由于多种原因造成的，例如计算机中的内核数以及数据库处理请求的能力。在测试过程中，您可以找到哪个快乐的数字。   

我总是从查看我有多少个内核和正在处理的工作类型开始。这项工作是否会被组织，平均持续多长时间。在Microsoft堆栈上，我发现每个内核三个活动线程似乎为大多数任务提供了最佳性能。我还不知道GO中的数字会是多少。  

您还可以为服务器需要处理的不同类型的工作创建不同的线程池。由于可以配置每个线程池，因此您可以花时间对服务器进行性能调整，以实现最大吞吐量。拥有这种类型的命令和控制以最大限度提高性能至关重要。  

在Go中，我们不创建线程，而是创建例程。例程功能类似于多线程函数，但Go管理操作系统级线程的实际使用。要了解有关GO中并发的更多信息，请查看以下[文档](http://golang.org/doc/effective_go.html#concurrency)  


我创建的包称为workpool和jobpool。他们使用channel和go例程构造来实现池化。  

#### Workpool 工作池  

此包创建一个go例程池，这些例程专用于处理发布到池中的工作。单个GO例程用于工作进行排队。队列例程提供安全的工作队列，跟踪队列中的工作量，并在队列已满时报告错误。  

将工作发布到队列中时阻止呼叫。这样，调用方就可以验证工作是否已排队。维护活动工作线程例程数的计数。  

以下是有关如何使用工作池的一些示例代码：  

```go
package main

import (
	"bufio"
	"fmt"
	"github.com/goinggo/workpool"
	"os"
	"runtime"
	"strconv"
	"time"
)

type MyWork struct {
	Name      string
	BirthYear int
	WP        *workpool.WorkPool
}

func (mw *MyWork) DoWork(workRoutine int) {
	fmt.Printf("%s: %d \n", mw.Name, mw.BirthYear)
	fmt.Printf("Q: %d, R: %d \n", mw.WP.QueuedWork(), mw.WP.ActiveRoutines())

	// Simulate some delay
	time.Sleep(100 * time.Millisecond)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	workPool := workpool.New(runtime.NumCPU(), 800)

	shutdown := false

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork{
				Name:      "A" + strconv.Itoa(i),
				BirthYear: i,
				WP:        workPool,
			}

			if err := workPool.PostWork("routine", &work); err != nil {
				fmt.Printf("ERROR: %s \n", err)
				time.Sleep(100 * time.Millisecond)
			}

			if shutdown {
				return
			}
		}
	}()

	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true

	fmt.Println("Shutting Down \n")
	workPool.Shutdown("name_routine")

}


```  

如果我们看一下main， 我们会创建一个线程池，其中要是用的例程数量屈居于我们机器上的内核数量。这意味着我们对每个内核都有一个历程。如果每个内核都繁忙，则无法再做任何工作。同样，性能测试讲确定这个数字应该是多少。第二个参数是队列的大小。在本例中，我已使队列足够大，以处理所有传入请求。  

MyWork 类型定义执行工作所需的状态。成员函数 DoWork 是必需的，因为它实现了 PostWork 调用所需的接口。若要将任何工作传递到线程池中，此方法必须按类型实现。  

DoWork 方法正在做两件事。首先，它显示对象的状态。其次，它报告队列中的项目数和 Go 例程的活动数。这些数字可用于确定线程池的运行状况和性能测试。


最后，我有一个 Go 例程，将工作发布到循环内的工作池中。与此同时，工作池正在为每个排队的对象执行 DoWork。最终，Go 例程完成，工作池继续工作。如果我们在任何时候按回车键，编程就会优雅地关闭。  


PostWork 方法可能会在此示例程序中返回错误。这是因为 PostWork 方法将保证工作被置于队列中，否则它将失败。失败的唯一原因是队列已满。设置队列长度是一个重要的考虑因素。  

#### Jobpool工作池

jobpool 包与 workpool 包类似，但有一个实现细节。此包维护两个队列，一个用于正常处理，一个用于优先级处理。优先级队列中的待处理作业始终在正常队列中的待处理作业之前得到处理。  

使用两个队列使 jobpool 比 workpool 更复杂一些。如果您不需要优先处理，那么使用工作池会更快、更高效   
以下是有关如何使用作业池的一些示例代码：  
```go 
package main

import (
    "fmt"
    "time"

    "github.com/goinggo/jobpool"
)

type WorkProvider1 struct {
    Name string
}

func (wp *WorkProvider1) RunJob(jobRoutine int) {
    fmt.Printf("Perform Job : Provider 1 : Started: %s\n", wp.Name)
    time.Sleep(2 * time.Second)
    fmt.Printf("Perform Job : Provider 1 : DONE: %s\n", wp.Name)
}

type WorkProvider2 struct {
    Name string
}

func (wp *WorkProvider2) RunJob(jobRoutine int) {
    fmt.Printf("Perform Job : Provider 2 : Started: %s\n", wp.Name)
    time.Sleep(5 * time.Second)
    fmt.Printf("Perform Job : Provider 2 : DONE: %s\n", wp.Name)
}

func main() {
    jobPool := jobpool.New(2, 1000)

    jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 1"}, false)

    fmt.Printf("*******> QW: %d AR: %d\n",
        jobPool.QueuedJobs(),
        jobPool.ActiveRoutines())

    time.Sleep(1 * time.Second)

    jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 2"}, false)
    jobPool.QueueJob("main", &WorkProvider1{"Normal Priority : 3"}, false)

    jobPool.QueueJob("main", &WorkProvider2{"High Priority : 4"}, true)
    fmt.Printf("*******> QW: %d AR: %d\n",
        jobPool.QueuedJobs(),
        jobPool.ActiveRoutines())

    time.Sleep(15 * time.Second)

    jobPool.Shutdown("main")
}


```  

在此示例代码中，我们创建两个工作线程类型结构。最好认为每个工人都是系统中的一些独立工作。  

我们主要创建一个包含 2 个作业例程的作业池，并支持 1000 个待处理作业。首先，我们创建 3 个不同的 WorkProvider1 对象并将它们发布到队列中，将优先级标志设置为 false。接下来，我们创建一个 WorkProvider2 对象并将其发布到队列中，将优先级标志设置为 true。  


由于作业池有 2 个例程，因此将首先处理排队的前两个作业。一旦其中一个作业完成，就会从队列中检索下一个作业。接下来将处理 WorkProvider2 作业，因为它已放置在优先级队列中。  

要获取工作池和工作池包的副本，请转到 github.com/goinggo   

一如既往，我希望这段代码能在某种程度上帮助你。


- 原文链接：https://www.ardanlabs.com/blog/2013/05/thread-pooling-in-go-programming.html



