### select语法常见问题

我正在一个已经在生产环境中运行的程序上测试新功能，突然代码表现得非常糟糕。我所看到的让我震惊，然后它为什么会发生就变得很明显了。我还一个竞争条件，只是等待成为一个问题。  

我是图提供代码和两个错误的简化版。  

- 清单1:  
```go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

var Shutdown bool = false

func debugSelect() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-sigChan:
			Shutdown = true
			continue
		case <-func() chan struct{} {
			complete := make(chan struct{})
			go LaunchProcessor(complete)
			return complete
		}():
			return

		}
	}

}

func LaunchProcessor(complete chan struct{}) {
	defer func() {
		close(complete)
	}()

	fmt.Printf("Start work \n")

	for count := 0; count < 5; count++ {
		fmt.Printf("Doing Work \n")
		time.Sleep(time.Second)
		if Shutdown == true {
			fmt.Printf("Kill Early \n")
			return
		}
	}
}

```  


此代码背后的思想是运行任务并终止。它允许操作系统请求程序提前终止。如果可以的话，我总是喜欢干净地关闭程序。  

示例代码创建一个绑定到操作系统信号的通道，并从终端窗口查找C。如果发生C，则Shutdown标识设置为true，程序继续返回到select语句。该代码还会生成一个执行工作的goroutine。该goroutine检查Shutdown标识，已确定程序是否需要提前终止。  

- **BUG Number1**  
请看这部分代码：  
```go
case <- func() chan struct{} {
	complete := make(chan struct{})
	go LaunchProcessor(complete)
	return complete
}():
```  
当我写这段代码时，我认为我是如此聪明。我认为动态执行一个函数来生成Go历程会很酷。它返回一个通道，选择者等待该通道被告知工作已经完成。当Goroutine完成时，它会关闭通道并终止进行。  

让我们运行程序：

- 完整运行输出：  
```bash
Start work 
Doing Work 
Doing Work 
Doing Work 
Doing Work 
Doing Work 
```  

正如预期的那样，程序将启动并声称Goroutine。Goroutine完成后，程序将终止。  

这一次我在程序运行时按C：

- 中途给退出信号：  
```bash
Start work 
Doing Work 
^CStart work 
Doing Work 
Kill Early 
Kill Early 
```  
**当我按C时，程序在此启动了GOroutine！！！**  

我以为与案例关联的函数只会执行一次。然后选择将等地通道向前移动。我不知道每次循环迭代回select语句时都会执行该函数。

为了修复代码，我需要从select语句中删除该函数，并在循环之外生成Goroutine：  
```go
func debugSelect2() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	complete := make(chan struct{})
	go LaunchProcessor(complete)

	for {
		select {
		case <-sigChan:
			Shutdown = true
			continue
		case <-complete:
			return

		}
	}
}
```  
现在我们运行程序时，我们会看到更好的结果：  

```bash
Start work 
Doing Work 
Doing Work 
^CKill Early 
```  

这一次，当我按C键时，程序会提前终止，并且不会在此生成另一个goroutine  

#### BUG2  

代码中还潜伏着第二个不太明显的错误，请看以下代码：  
```go
var Shutdown bool = false

if whatSig == syscall.SIGINT {
	Shutdown = true
}

if Shutdown == true {
	fmt.Printf("Kill Early \n")
	return
}

```  

该代码使用包级别变量来之时正在运行的goroutine，在命中C时关闭。每次我按C键时代码都在工作，那么会出现什么错误呢？  

首先，让我们针对代码运行竞争检测器：  
```bash
go build -race -o select_debug_main ./main.go
./select_debug_main
```  
当它运行时，我再次点击C：  
```bash
Start work 
Doing Work 
Doing Work 
^C==================
WARNING: DATA RACE
Write at 0x000001207208 by main goroutine:
  main.main()
      /Users/slice/GolandProjects/learn-go/golang/example/ch/main.go:26 +0x10a

Previous read at 0x000001207208 by goroutine 8:
  main.LaunchProcessor()
      /Users/slice/GolandProjects/learn-go/golang/example/ch/main.go:45 +0x124
  main.main.func1()
      /Users/slice/GolandProjects/learn-go/golang/example/ch/main.go:21 +0x33

Goroutine 8 (running) created at:
  main.main()
      /Users/slice/GolandProjects/learn-go/golang/example/ch/main.go:21 +0xfc
==================
Kill Early 
Found 1 data race(s)

```  
我对Shutdown标志的使用出现在竞争检测器上。这是因为我有两个Go历程会试图以不安全方式访问变量。  

我最初不保护对变量的访问是实用的，但是错误的。我认为由于该变量仅用于必要时关闭程序，因此我不在乎脏度。如果碰巧在微秒的搜索范围内，在写入变量和读取变量之间，如果发生了脏读，我会在下一个循环中再次捕获它。没有造成伤害吧，对吧？**为什么要为这样的事情添加复杂的通道或锁定代码？**  

嗯，有一个小东西叫做Go内存模型：

[http://golang.org/ref/mem](http://golang.org/ref/mem)   

Go内存模型不保证读取Shutdown变量的Go进程会看到主进程的写入操作。祝进程对Shutdown变量的写入永远不会写回主内存是有效的。这是因为祝进程从不读取Shutdown变量。  

这在今天不会发生，但随着Go编译器变得越来越复杂，它可能会决定完全消除对Shutdown变量的写入。Go内存模型允许此行为。此外我们不希望代码无法通过竞争检测，这知识不好的做法，即使处于实际原因也是如此。  

以下是代码的最终版本，修复了所有的错误：  

```go
func debugSelectMain() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	complete := make(chan struct{})
	go LaunchProcessor(complete)

	for {
		select {
		case <-sigChan:
			atomic.StoreInt32(&Shutdown, 1)
			continue
		case <-complete:
			return

		}
	}
}

func LaunchProcessor(complete chan struct{}) {
	defer func() {
		close(complete)
	}()

	fmt.Printf("Start work \n")

	for count := 0; count < 5; count++ {
		fmt.Printf("Doing Work \n")
		time.Sleep(time.Second)
		if atomic.LoadInt32(&Shutdown) == 1 {
			fmt.Printf("Kill Early \n")
			return
		}
	}
}
```  

我更喜欢实用if语句来检查是否设置了Shutdown标志，以便我可以根据需要散布该代码。次解决方案将Shutdown标识从布尔值更改为int32，并使用原子函数Store和Load。  

在主进程中，如果检测到C，则Shutdown标志会安全地从0改为1.在LaunchProcessor Goroutine中，将Shutdown标志的值与1进行对比。如果该条件为true，则返回goroutine。  

有时令人惊讶的是，像这样的简单程序也可以包含一些陷阱。你刚开始的时候可能从未想过或意识到的事情。尤其是当代码似乎总是有效时。  

- 原文: https://www.ardanlabs.com/blog/2013/10/my-channel-select-bug.html


