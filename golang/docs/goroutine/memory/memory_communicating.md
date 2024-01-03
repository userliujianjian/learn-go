## 通过通信共享内存  

#### 简介  
传统的线程模式（例如：通常用于编写java、C++和python程序）要求程序员使用共享内存在县城之间进行通信。通常，共享数据结构受锁保护，线程将争用这些锁以访问数据。在某些情况下，通过使用线程安全的数据结构（例如Python队列）可以简化此操作。  
<br>

Go的并发原语，goroutine和channels, 提供了一种优雅而独特的并发软件结构方法。（这些概念有一段有趣的历史，始于C.A.R.的Communicating Sequential Processes.）Go鼓励使用通道在goroutine之间传递数据的引用，而不是显式使用锁来调节对共享数据的访问。这种方法可确保在给定时间只有一个goroutine可以访问数据。这个概念总结在文档Effective Go中（任何Go程序员必读书）

> Tips: 不哟啊通过共享内存进行交流，通过交流共享内存。（CSP核心思想）

考虑一个轮询URL列表的程序。在传统的线程环境中，可以这样构建其数据结构：  
```go
type Resource struct {
	url        string
	polling    bool
	lastPolled int64
}

type Resources struct {
	data []*Resource
	lock *sync.Mutex
}
```

然后，轮询器函数（其中许多将在单独的线程中运行）可能如下所示：  
```go
func Poller(res *Resources) {
	for {
		//	 get the least recently-polled Resource
		// and mark it as being polled
		res.lock.Lock()
		var r *Resource
		for _, v := range res.data {
			if v.polling {
				continue
			}

			if r == nil || v.lastPolled < r.lastPolled {
				r = v
			}

		}

		if r != nil {
			r.polling = true
		}
		res.lock.Unlock()

		if r == nil {
			continue
		}

		// pool the url

		// update the resource's polling and lastPolled
		res.lock.Lock()
		r.polling = false
		r.lastPolled = int64(time.Nanosecond)
		res.lock.Unlock()
	}
}
```

此函数大约有一页长，需要更多细节才能完成。它甚至不包括url轮询逻辑（它本身只有几行），也不会优雅地处理耗尽资源池的问题。  
<br>

让我们来看一下使用go惯用语实现的相同功能。再次示例中，轮询器是一个函数，用于从输入通道接收要轮询的资源，并在完成后将其发送到输出通道。  

```go
type Source string

func Poller2(in, out chan *Source) {
	for r := range in {
		// poll the URL

		// send the processed Resource to out
		out <- r
	}
}
```  

前面这些例子中微妙逻辑明显确实，我们的Source数据结构不再抱函记录数据。事实上，剩下的只是重要的部分。这应该让你对这些简单语言功能的强大有所了解。  

上面的代码片段有很多遗漏。有关这些想法的完整惯用Go程序演练，请参阅[Codewalk](https://go.dev/doc/codewalk/sharemem/)通过通信共享内存。  

- 参考文章：
- [原文：通过通信共享内存 VS 通过共享内存来通信？](https://go.dev/blog/codelab-share)  

