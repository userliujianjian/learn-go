## Go编程中的线程池  

在Go中工作了一段时间后，我学会了如何使用无缓冲通道来构建goroutines池。我比这篇文章中实现的更好。话虽如此，这篇文章所描述的内容仍然有价值（https://github.com/goinggo/work）。

### 介绍  
在我的服务器开发领域，线程池一直在Microsoft堆栈上构建健壮代码的关键。微软在.Net中失败了，它为每个进程提供了一个包含数千个线程的线程池，并认为它们可以在运行时管理并发性。我很早就意识到这永远行不通。至少对于我正在开发的服务器来说不是。  

当我使用Win32 API在C/C++中构建服务器时，我创建了一个抽象的IOPC的类，为我提供可以将工作发布到的线程池。这一直工作的很好，因为我可以定义池中的线程数和兵法几杯（在任何给定时间允许活动的线程数）。我为我的所有C#开发移植了这段代码。如果你想了解更多信息，我几年前写过一篇文章（http://www.theukwebdesigncompany.com/articles/iocp-thread-pooling.php） 。使用IOCP为我提供了所需的性能和灵活性。顺便说一句，.Net线程池在底层使用IOCP。  

线程池的想法相当简单。工作进入服务器并需要进行处理。大多数工作本质上是异步的，但并非必须如此。很多时候，工作是从套接字或内部例程中完成的。线程池将工作排队，然后从池中分配一个线程来执行该工作。工作按照收到的顺序进行处理。该池提供了高效执行工作的良好模式。每次需要处理工作时生成一个新线程给操作系统带来沉重的负载并导致严重的性能问题。  
  
那么线程池性能是如何调优的呢？您需要确定每个池应包含的线程数，以便最快地完成工作。当所有例程都忙于处理工作时，新工作将保持排队状态。您想要这样做是因为某些时候有更多例程处理工作会减慢速度。这可能有多种原因，例如计算机中的内核数量以及数据库处理请求的能力。在测试过程中你可以找到那个快乐的数字。  

我总是首先查看我有多少个核心以及正在处理的工作类型。这项工作是否会被阻止以及平均持续多久。在Microsoft堆栈上，我发现每个核心三个活动线程似乎可以为大多数任务带来最佳的性能。我还不知道GO中的数字是多少。  

您还可以为服务器需要处理的不同类型的工作创建不同的线程池。鱿鱼每个线程池都可以配置，因此您可以花时间调整服务器性能以获得最大的吞吐量。拥有这种类型的命令和控制来最大限度的提高性能至关重要。  

在Go中，我们不创建线程池，而是创建例程。这些例程的功能类似于出多线程函数，但GO管理操作系统级线程的实际使用。要了解有关Go中并发性更多信息，请查看此文档：http://golang.org/doc/effective_go.html#concurrency 。  

我创建的包称为工作池和作业池。它们使用通道和Go例程结构来实现池化。  

### Wrokpool工作池    

该包创建了一个go例程池，专门用于处理发布到池中的工作。使用单个go例程对工作进行排队。队列例程提供安全的工作排队、跟踪队列中的工作量并在队列已满时报告错误。  

将工作发布到队列中是一个阻塞调用。这样调用者就可以验证工作是否已排队。维护活跃工人例程的数量计数。  

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
	fmt.Printf("%s : %d \n", mw.Name, mw.BirthYear)
	fmt.Printf("Q: %d, R: %d \n", mw.WP.QueuedWork(), mw.WP.ActiveRoutines())

	// Simulate some delay
	time.Sleep(100 * time.Millisecond)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	pool := workpool.New(runtime.NumCPU(), 800)
	shutdown := false // Race condition, sorry

	go func() {
		for i := 0; i < 1000; i++ {
			work := MyWork{
				Name:      "A" + strconv.Itoa(i),
				BirthYear: i,
				WP:        pool,
			}
			if err := pool.PostWork("routine", &work); err != nil {
				fmt.Printf("ERROR: %s \n", err)
				time.Sleep(100 * time.Millisecond)
			}
			if shutdown == true {
				return
			}
		}
	}()

	fmt.Println("Hit any key to exit")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	shutdown = true
	fmt.Println("Shutting Down")
	pool.Shutdown("routine")
}

```  
如果我们查看main，我们会创建一个线程池，其中要使用的例程数量基于机器上的核心数量。这意味着我们每个核心都有一个例程。如果每个核心都很忙，您就无法再做任何工作。同样，性能测试将确定这个数字应该是多少。第二个参数是队列的大小。在本例中，我已使队列足够大以处理所有传入的请求。  

Mywork类型定义了我执行工作所需的状态。成员函数DoWork是必需的，因为它实现了PostWork调用所需的接口。要将任何工作传递到线程池中，必须由类型实现此方法。  

DoWork方法正在做两件事。首先它显示对象的状态。其次，它报告队列中的项目数量和活动的Go例程数量。这些数字可用于确定线程池的运行情况和性能测试。  

最后我有一个Go例程，将工作发送到循环内的工作池中。与此同时，工作池正在为每个排队的对象执行DoWork。最终Go例程完成，工作池继续执行其工作。如果我们随时按回车键，程序就会正常关闭。  

在此示例程序中，PostWork方法可能会返回错误。这是因为PostWork方法将保证工作被放入队列中，否则将会失败。失败的唯一原因是队列已满。设置队列长度是一个重要的考虑因素。  

### Jobpool作业池  

除了一个实现细节之外，jobpool包与workpool包类似。该包维护两个队列，一个用于正常处理，一个用于优先处理。优先级队列中的待处理作业始终优先于普通作业得到处理。  

两个队列的使用使得作业池比工作池稍微复杂一些。如果您不需要优先级处理，那么使用工作池将会更快、更高效。

以下是有关如何使用作业池的一些示例代码：  
```go
package main

import (
	"fmt"
	"github.com/goinggo/jobpool"
	"time"
)

type WorkProvider1 struct {
	Name string
}

func (wp *WorkProvider1) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 1 : started: %s \n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 1 : Done: %s \n", wp.Name)
}

type WorkProvider2 struct {
	Name string
}

func (wp *WorkProvider2) RunJob(jobRoutine int) {
	fmt.Printf("Perform Job : Provider 2 : started: %s \n", wp.Name)
	time.Sleep(2 * time.Second)
	fmt.Printf("Perform Job : Provider 2 : Done: %s \n", wp.Name)
}

func main() {
	jobPool := jobpool.New(2, 1000)
	_ = jobPool.QueueJob("main", &WorkProvider1{"Normal Priority: 1"}, false)
	fmt.Printf("***************> QW: %d AR: %d \n", jobPool.QueuedJobs(), jobPool.ActiveRoutines())

	time.Sleep(1 * time.Second)
	jobPool.QueueJob("main", &WorkProvider1{"normal Priority: 2"}, false)
	jobPool.QueueJob("main", &WorkProvider1{"normal Priority: 3"}, false)

	jobPool.QueueJob("main", &WorkProvider2{"Normal Priority: 4"}, true)
	fmt.Printf("***************> QW: %d AR: %d \n", jobPool.QueuedJobs(), jobPool.ActiveRoutines())
	time.Sleep(15 * time.Second)

	jobPool.Shutdown("main")
}

```  

在此示例代码中，我们创建了两个工作类型结构。最好认为每个工人都是系统中的一些独立工作。  

主要是我们创建了一个包含2个作业例程的作业池，并支持1000个待处理作业。首先我们创建3个不同的WorkProvider1对象并将它们放入队列中，并将优先级标志设置为false。接下来，我们创建一个WorkProvider2对象并将其发布到队列中，并将优先级标志设置为true。  

鱿鱼作业池有2个例程，因此将首先处理排队的前两个作业。一旦其中一个作业完成就会从队列中检索下一个作业。加下来将处理WorkProvider2作业，因为它被放置在优先级队列中。  

要获取工作池和作业池包的副本，请访问github.com/goinggo

一如既往，我希望这段代码可以在一些小方面对您有帮助。  





