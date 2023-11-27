### scheduler调度器介绍
Go程序的执行由两层组成： Go Program, Runtime, 即用户程序和运行时。 它们之间通过函数调用实现内存管理、channel通信、goroutines创建等功能。用户程序进行的系统调用都会被Runtime拦截，以此来帮助它进行调度以及垃圾回收相关工作。
![总揽全图](../img/schedule-1.png)

- 为什么要scheduler
	Go scheduler可以说是Go运行时的一个重要的部分。Runtime维护所有的goroutines，并通过scheduler来进行调度。Goroutines和threads是独立的，但是goroutines要依赖threads才能执行。 Go程序执行的高效和scheduler的调度是分不开的。

- **scheduler底层原理**。
	实际上操作系统角度，所有程序都是在执行多线程。将goroutines调度到线程M上执行，仅仅是runtime层面的一个概念，在操作系统之上的层面。  
	有三个基础结构来实现goroutines调度： GMP。
	runtime起始时会启动一些G：垃圾回收的G，执行调度的G，运行用户代码的G；并且会创建一个M用来开始G的运行。随着时间的推移，更多的G会被创建出来，更多的M也会被创建出来。  

	当然在Go早期版本并没有P这个结构体，m必须从一个全局队列里获取要运行的g 因此需要获取一个全局的锁，当并发量大的时候，锁就成了瓶颈。后来在大神Dmitry Vyokov的实现里，加上了P结构体。每个P自己维护一个处于Runnable状态的G的队列，解决了原来的全局锁问题。

- **GOscheduler的核心思想**：
	- reuse threads;
	- 限制同时运行（不包含阻塞）的线程数为N，N等于CPU的核心数；
	- 线程私有的runqueues，并且可以从其他线程stealing goroutine来运行，线程阻塞后，可以将runqueues传递给其他线程。

- 为什么要P这个组件，志杰吧runqueues放到M不行嘛？
	- 当一个线程阻塞的时候，将和它绑定的P上的goroutines可以转移到其他线程。P为M提供上下文，该去执行哪个G

Go scheduler会启动一个后台线程sysmon，用来检测长时间（超过10ms）运行的goroutine，将其调度到global runqueues。这是一个全局的runqueue，优先级比较低，以示惩罚。

GPM都说完了，有两个重要组件还没提到： 全局可运行队列（GRQ）和本地可运行队列（LRQ）。 LRQ存储本地（也就是具体的P）的可运行的goroutine，GRQ存储全局的可运行goroutine，这些goroutine还没有分配到具体的P。

![GRQ-LRQ](../img/schedule-2.png)

Go scheduler是Go runtime的一部分，它镶嵌在Go程序中和Go程序一起运行。因此它运行在用户空间，在kernel的上一层。和OS scheduler抢占调度(preemptive)不一样，Go scheduler采用协作式调度(cooperating)。

协作式调度一半会由用户设置调度点，如python中的yield会告诉os scheduler可以将我调度出去了。

但是由于在Go语言里，goroutine调度的事情是由Go runtime来做，并非由用户控制，所以我们依然可以讲Go scheduler看成是抢占式调度，因为用户无法预测调度器下一步的动作是什么。

和线程类似，goroutine的状态也是三种（简化版）：
- waiting: 等待状态，goroutine在等待某件事的发生。例如等待网络数据、硬盘；调用操作系统API；等带内存同步访问条件ready，如atomic，mutexes
- Runnable： 就绪状态，只要给M我就可以运行
- Executing: 运行状态。goroutine在M上执行命令，这是我们想要的。

下面这张GPM全局的运行示意图见的比较多，可以留着。

![GPM-struct](../img/schedule-3.png)