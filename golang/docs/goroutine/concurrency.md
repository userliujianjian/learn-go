## 并发模型之goroutine的使用时机

#### 简介
通常，Go因其并发特性而被选为项目语言。Go团队不遗余力的使go中的并发性成本低廉（就硬件资源而言）和性能，但是可以使用go的并发特性来别写既不高性能也不可靠的代码。在剩下的时间里，我想给你一些建议，以避免go的并发功能带来的一些陷阱。  

Go具有一流的并发支持，包括通道和select and go语句。如果你从书本或培训课程中正式学习了go，你可能已经注意到并发部分始终是你最后要介绍的部分之一。这次研讨会也不例外，我选择最后介绍并发性，好像它是Go程序员应该掌握的常规技能的某种补充。  

这里有一个二分法；Go的主要特点是简单、轻量级并发模型。 作为一个产品，我们的语言几乎只能靠这个功能来推销自己。另一方面，有一种说法认为并发实际上并不那么容易使用，否则作者不会把它作为他们书的最后一张，我们也不会遗憾的回顾过去的努力。  

本届讨论优质的使用Go并发功能的一些陷阱

#### **Keep yourself busy or do the work yourself（自己的事情自己做）**。
- 下面这段代码有什么问题
```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func start1() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, GopherCon SG")
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		
	}

}
```
该程序使用net库中的http，提供简单的web服务。服务是通过goroutine启动的，同时for阻塞start1函数，它会无在无限循环中浪费CPU。
goroutine跟主进程之间没有任何约束，goroutine上web服务挂了，主程序是无感知的，不会退出。   

由于go运行时大多是合作调度的，因此该程序将在单个CPU上徒劳无功旋转，最终可能被实时锁定。  

- 如何解决这个问题呢？
```go
func start2() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, GopherCon SG")
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		runtime.Gosched()
	}

}
```
这可能看起来很傻，但这是我在野外看到的常见解决方案。这是不了解潜在问题的症状。  

现在如果你有一些经验，你可能会这样写：
```go
func start3() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, GopherCon SG")
	})

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	select {}

}
```
**空select语句将永远阻塞**。这是一个有用的属性，因为现在我们不仅仅为了调用runtime.GoSched() 然而，我们只是在治疗症状，并不是病因。

我想给你们介绍一种方案，希望你们已经想到了，与其在goroutine中运行 http.ListenAndServe, 让我们面临如何处理goroutine问题，不如简单地在goroutine本身上运行 http.ListenAndServe   
```go
func start4() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, GopherCon SG")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
```
- 所以这是我第一条建议：如果你的goroutine在另外一个人那里得到结果之前无法取得进展，那么通常只自己做工作比委托它更简单。
- 这通常消除了goroutine传回其发起方所需的大量状态跟踪和管道操作。
- 许多 Go 程序员过度使用 goroutines，尤其是在他们刚起步的时候。**与生活中的所有事情一样，适度是成功的关键**。



#### 思考
> 创建一个简单的web服务和一个监控web的影子（类似于web服务）两个服务，如何创建？   
> Tip: 两个服务需要同时关闭，有一个关闭另一个也会随着关闭


参考文献：
https://dave.cheney.net/practical-go/presentations/qcon-china.html#_concurrency





