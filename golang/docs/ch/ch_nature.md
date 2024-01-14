### Go中Channel的本质

#### 介绍  

在上一篇文章[《Concurrency， Goroutines and GOMAXPROCS》](../goroutine/goroutines-and-gomaxprocs.md)中，为讨论通道奠定了基础。我们讨论了并发，以及goroutine时如何发挥作用的。有了这个基础，我们现在可以理解通道的本质，以及如何使用它来同步goroutine，以安全、不易出错和有趣的方式共享资源。  

#### 什么是channel  
通道是类型安全的消息队列，具有控制任何尝试接收或发送的goroutine行为的能力。通道充当两个goroutine之间的管道，并将同步通过它传递任何资源的交换。正是通道由控制goroutines交互的能力创建了同步机制。当创建没有容量的通道时，它称为无缓冲通道。反过来，使用容量创建的通道成为缓冲通道。  

#### 无缓冲通道  
无缓冲通道没有容量，因此需要两个goroutine都准备好后才能进行任何交换（*发送发生在接收完成之前*）。当goroutine尝试将资源发送到无缓冲通道并且没有goroutine等待接收资源时，该通道将锁定发送的goroutine并使其等待。当goroutine尝试从无缓冲通道接收，并且没有goroutine等待发送资源时，该通道锁定接收的goroutine并使其等待。  

![goroutine-swap](../img/ch-nature-1.png)  

再上图中，我们看到了两个goroutine使用无缓冲通道进行交换的示例。在步骤1中，两个goroutine接近通道，步骤二中，左侧的goroutine将手伸入通道或执行发送。此时该goroutine将被锁定在通道中，知道交换完成。然后在第三步中，右侧goroutine将他的手放入通道或执行接收。该goroutine也被锁定在通道中，直到交换完成。在第4步和第5步中进行交换。然后第6步中，两个goroutine都可以自由地移开它们的手并继续前进。  

*同步是发送和接收之间的交互所固有的。一个没有另一个就不可能发生。无缓冲通道的本质是保证同步*  

#### 缓冲通道  
缓冲通道具有容量，因此其行为可能略有不同。当goroutine尝试将资源发送到缓冲通道并且通道已满时，该通道将锁定goroutine并使其等待，知道缓冲区可用。如果通道中有空间，则可以立即进行发送，并且goroutine可以继续前行。当goroutine尝试从缓冲通道接收并切缓冲通道为空时，通道将锁定goroutine并使其等待资源发送完毕。  

![ch-buffer-swap](../img/ch-nature-2.png)  

再上图中，我们看到两个goroutine独立地在缓冲通道中添加和删除项目的例子。在步骤1中，右侧的goroutine是从通道中删除资源或执行接收。在步骤2中，右侧的goroutine可以独立于左侧的goroutine删除资源，从而向通道添加新资源。在步骤3中，两个goroutine同时在通道中添加和删除资源。在步骤4中，两个goroutines都完成。  

同步仍然发生在接收和发送的交互中，但是当队列具有缓冲区可用时，发送不会锁定。当有东西要从通道接收时，接收不会锁定。因此，**如果缓冲区已满或没有可接收的内容时，则缓冲通道的行为将与无缓冲通道非常相似。**  

#### 接力赛（relay race）  

如果你曾经看过Tina斤比赛，你可能看过接力赛。在接力赛中，有4名运动员作为一个团队尽可能快地在赛道上奔跑。比赛的关键是每支队伍一次能有一名跑步者跑步。拿着接力棒的跑步者是唯一允许跑步的人，而接力棒在跑步者与跑步者之间的交换对于赢得比赛至关重要。  

让我们构建一个示例程序，它使用四个goroutine和一个通道来模拟接力赛。goroutines将是比赛中的跑步者，通道将用于在每个跑步者之间交换接力棒。这是一个经典的例子，说明资源如何在goroutines之间传递，以及通道如何控制欲知交互的goroutines的行为。  

```go
package main

import (
	"fmt"
	"time"
)

func swapState() {
	// create an unbuffered channel
	baton := make(chan int)

	// First runner to his mark
	go Runner(baton)

	// start the race
	baton <- 1

	// Give the runners time to race
	time.Sleep(500 * time.Millisecond)
}

func Runner(baton chan int) {
	var newRunner int

	// wait to receive the baton
	runner := <-baton

	// start running around the trace
	fmt.Printf("Runner %d Running With Baton \n", runner)

	// new runner to the line
	if runner != 4 {
		newRunner = runner + 1
		fmt.Printf("Runner %d To the line \n", newRunner)
		go Runner(baton)
	}

	// running around the track
	time.Sleep(100 * time.Millisecond)

	// is the race over
	if runner == 4 {
		fmt.Printf("Runner %d Finished, Race over \n", runner)
		return
	}

	// exchange th baton for the next runner
	fmt.Printf("Runner %d Exchange with runner %d \n", runner, newRunner)
	baton <- newRunner

}

```  

当我们运行示例程序时，我们得到以下输出：

```bash
Runner 1 Running With Baton 
Runner 2 To the line 
Runner 1 Exchange with runner 2 
Runner 2 Running With Baton 
Runner 3 To the line 
Runner 2 Exchange with runner 3 
Runner 3 Running With Baton 
Runner 4 To the line 
Runner 3 Exchange with runner 4 
Runner 4 Running With Baton 
Runner 4 Finished, Race over 
```  

程序开始创建一个无缓冲通道：  


```go
// create an unbuffered channel
baton := make(chan int)		
```  

使用无缓冲通道会强制 goroutine 同时准备好进行接力棒交换。两个 goroutine 都准备就绪的需要创建了有保证的同步。  

如果我们看一下 main 函数的其余部分，我们会看到为比赛中的第一个跑步者创建一个 goroutine，然后将接力棒交给该跑步者。此示例中的接力棒是在每个运行器之间传递的整数值。该示例使用睡眠让比赛在主终止和结束程序之前完成：  

```go
// create an unbuffered channel
	baton := make(chan int)

	// First runner to his mark
	go Runner(baton)

	// start the race
	baton <- 1

	// Give the runners time to race
	time.Sleep(500 * time.Millisecond)
```  

如果我们只关注 Runner 功能的核心部分，我们可以看到接力棒交换是如何发生的，直到比赛结束。Runner 函数作为比赛中每个跑步者的 goroutine 启动。每次启动新的 goroutine 时，通道都会传递到 goroutine 中。通道是交换的管道，因此当前运行者和等待下一个运行者需要引用该通道：  

```go
func Runner（baton chan int）
```  


每个跑步者做的第一件事就是等待接力棒交换。这是通过通道上的接收来模拟的。接收会立即锁定 goroutine，直到接力棒被发送到通道中。一旦接力棒被送入通道，接收将释放，goroutine 将模拟下一个跑步者在赛道上冲刺。如果第四名选手正在跑步，则不会有新的选手参加比赛。如果我们仍然在比赛中途，就会为下一位跑步者启动一个新的 goroutine。  

```go
// wait to receive the baton
	runner := <-baton

	// start running around the trace
	fmt.Printf("Runner %d Running With Baton \n", runner)

	// new runner to the line
	if runner != 4 {
		newRunner = runner + 1
		fmt.Printf("Runner %d To the line \n", newRunner)
		go Runner(baton)
	}
```  


然后我们睡觉来模拟跑步者在跑道上跑所需的一些时间。如果这是第四位跑步者，则 goroutine 在睡眠后终止，比赛结束。如果没有，则在发送到通道时进行接力棒交换。有一个 goroutine 已经锁定并等待此交换。接力棒一送入通道，就进行交换，比赛继续进行：

```go
// running around the track
	time.Sleep(100 * time.Millisecond)

	// is the race over
	if runner == 4 {
		fmt.Printf("Runner %d Finished, Race over \n", runner)
		return
	}

	// exchange th baton for the next runner
	fmt.Printf("Runner %d Exchange with runner %d \n", runner, newRunner)
	baton <- newRunner
```

#### 结论  

该示例展示了一个真实世界的事件，即跑步者之间的接力赛，以模仿实际事件的方式实现。这是频道的美妙之处之一。代码的流动方式模拟了这些类型的交换在现实世界中是如何发生的。  

现在我们已经了解了无缓冲通道和缓冲通道的性质，我们可以看看可以使用通道实现的不同并发模式。并发模式允许我们在模拟真实世界计算问题的 goroutines 之间实现更复杂的交换，例如信号量、生成器和多路复用器。  


- 参考文章：https://www.ardanlabs.com/blog/2014/02/the-nature-of-channels-in-go.html




