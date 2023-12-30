## 并发模型之开启和关闭（Never Start a goroutine without knowning when it will stop）
#### 简介：
前面的示例显示了在不需要goroutine时使用goroutine。但使用Go的驱动原因之一是该语言提供的并发功能。事实上，许多情况下，您希望利用应硬件中可用的并行性。为此，您必须使用goroutines

这个简单的应用程序在两个不同的端口上提供http流量，端口8080用于应用程序流量，端口8001用于访问端点 /debug/pprof

```go
package main

import (
	"fmt"
	"net/http"
)

// never start a goroutine without knowing when it will stop(永远不要再不知道它合适停止的情况下启动一个goroutine)

func startPprof() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello, QCon!")
	})

	go http.ListenAndServe(":8001", http.DefaultServeMux)
	http.ListenAndServe(":8080", mux)

}
```
虽然这个程序不是很复杂，但它代表了实际应用程序的基础。  
应用程序目前存在一些问题，这些问题会随着应用程序的增长而显现出来，因此现在让我们解决其中的一些问题。

```go
func serveApp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello, QCon!")
	})

	http.ListenAndServe(":8080", mux)

}

func serveDebug() {
	http.ListenAndServe(":8001", http.DefaultServeMux)
}

func startMain() {
	go serveDebug()
	serveApp()
}
```
通过将serveApp、serveDebug将程序分解，将它们于main.main解耦。我们还遵循了上面的建议，并确保serveApp将其并发serveDebug留给调用方。  

但是这个程序存在一些可操作性问题。如果serveApp返回，则main.main返回，导致程序关闭并由您使用的任何进程管理重新启动。

> TIP: 正如go中的函数将并发权交给调用者一样，应用程序应该将监视其状态，并在失败时重新启动他们的工作留给调用它们的程序。不要让程序负责自行重新启动，这是最好从应用程序外部处理的过程。  

但是，在单独的goroutine中运行，serverDebug如果它们返回，则goroutine将推出，而程序的其余部分继续运行。您的操作人员不会很高兴的发现，当从程序中无法调取统计信息时才发现停止了，/debug因为处理程序很久以前就停止了工作。

我们想要确保的是，**如果任何负责为该应用程序提供服务的goroutine停止，我们将关闭该应用程序**  

```go
func serveApp2(){
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request){
		fmt.Fprintln(resp, "Hello QCon!")
	})
	
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func serveDebug2(){
	if err := http.ListenAndServe(":8001", http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}

func startMain2(){
	go serveApp2()
	go serveDebug2()
	select {}
}
```
现在serveApp2和serveDebug2返回错误，ListenAndServe并在需要时调用log.Fatal。因为两个处理程序都在goroutines中运行，所以我们将主goroutine阻塞在select{}。

- 这种方法存在许多问题：
	- 如果ListenAndServer 返回错误nil，则不会被调用，log.Fatal并且该端口上的HTTP服务将关闭而不停止应用程序。
	- log.Fatal将无条件退出程序，调用os.Exit;defers不会被调用，其他goroutines不会被通知被关闭，程序只会停止。这使得这些函数编写测试变得困难。

> Tip: log.Fatal只有在main函数或者init函数中使用。

- 我们真正想要的是将发生的任何错误回传给goroutine的发起者，以便它可以知道goroutine停止的原因，可以优雅的关闭进程。  
```go
func serveApp3() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello QCon!")
	})
	return http.ListenAndServe(":8080", mux)
}

func serveDebug3() error {
	return http.ListenAndServe(":8001", http.DefaultServeMux)
}

func startMain3() {
	done := make(chan error, 2)
	go func() {
		done <- serveApp3()
	}()

	go func() {
		done <- serveDebug3()
	}()

	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			fmt.Printf("error: %v", err)
		}
	}
}
```
我们可以使用通道来收集goroutine的返回状态。通道的大小等于我们要管理的goroutine的数量，这样发送到done通道就不会阻塞，否则会阻塞goroutine的关闭，导致它泄漏。  

由于无法安全地关闭done通道，因此我们不能使用for range习惯用法来循环通道，直到所有 goroutine 都已报告为止，而是循环我们启动的 goroutine 数量，这等于通道的容量。  

现在我们有办法等待每个goroutine完全退出并记录它们遇到的任何错误。所需要的只是一种将关闭信号从第一个推出的goroutine转发给其他goroutine的方法。  

事实证明，要求一个http.server关闭有点复杂，所以我们将该逻辑分几位一个辅助函数。该serve帮助器采用一个地址和http.Handler,类似于http.ListenAndServe, 以及一个stop我们用来出发该ShutDown方法的通道。

```go
func serve(addr string, handler http.Handler, stop <-chan struct{}) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-stop
		s.Shutdown(context.Background())
	}()

	return s.ListenAndServe()

}

func serveAppMaster(stop <-chan struct{}) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "hello QCon!")
	})

	return serve(":8080", mux, stop)
}

func serveDebugMaster(stop <-chan struct{}) error {
	return serve(":8001", http.DefaultServeMux, stop)
}

func mainMaster() {
	done := make(chan error, 2)

	stop := make(chan struct{})

	go func() {
		done <- serveAppMaster(stop)
	}()
	
	go func() {
		done <- serveDebugMaster(stop)
	}()
	
	var stopped bool
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			fmt.Printf("error: %v", err)
		}
		if !stopped {
			stopped = true
			close(stop)
		}
	}
}

```

现在，每次我们在done通道上收到一个值是，我们都会关闭该stop通道，这会导致所有等待该通道goroutine关闭它们的http.Serve. 这反过来将导致所有剩余的 ListenAndServe goroutine返回。一旦我们启动的所有goroutine都停止了，main.main函数就会返回并且进程会干净地停止。 

> Tip: 自己编写这个逻辑是重复且微妙的。考虑像这个包这样的东西，[https://github.com/heptio/workgroup](https://github.com/heptio/workgroup) 它将为您完成大部分工作。


#### 参考文献：
https://dave.cheney.net/practical-go/presentations/qcon-china.html#_concurrency




