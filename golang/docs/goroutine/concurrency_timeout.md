### GO并发模式：超时，继续前进

并发变成有自己的习语。一个很好的例子是超时。虽然Go的通道不直接支持它们，但它们很容易实现。假设我们想从通道ch接收，但最多想等一秒钟才能达到值。我们将首先创建一个信令通道，并启动一个在通道上发送之前休眠的goroutine：  

```go
timeout := make(chan bool, 1)
go func(){
	time.Sleep(1 * time.Second)
	timeout <- true
}
```

然后，我们可以使用语句select或从ch timeout接收。如果一秒钟后没有任何结果ch，则选择超时情况，并放弃从ch读取的尝试。
```go
select{
case <- ch:
	// a read from ch has occurred
case <- timeout:
	// the read from ch has timeout
}
```  

timeout通道缓冲了1个值的空间，允许超时goroutine发送到通道，然后退出。goroutine不知道（或不关心）是否接收到值。这意味着，如果ch在达到超时之前发生接收，则goroutine不会永远挂起。该timeout通道最终将被垃圾收集器释放。

在这个例子中，我们曾经使用time.Sleep演示过goroutines和channels机制。在实际程序中，您应该使用[time.After], 一个返回通道并在指定持续时间后在该通道上发送的函数。  

让我们来看看这种模式的另一个变体。在此示例中我们有一个同事从多个复制数据库读取数据的程序。该程序只需要一个答案，并且它应该接受第一个到达的答案。  

该函数Query采用一段数据库连接和一个query字符串。它并行查询每个数据库并返回收到的第一个响应：  
```go
func Query(conns []Conn, query string) Result {
	ch := make(chan Result)
	for _, conn := range conns{
		go func(c Conn){
			select {
			case ch <- c.DoQuery(query):
			default:
			}
		}(conn)
	}
	return <- ch
}

```  
在此示例中，闭包执行非阻塞发送，这是通过在select带有case的语句中使用发送操作来实现的default。如果发送无法立即完成，则将选择默认情况。使发送成为非阻塞可以保证循环中启动的任何goroutine都不会挂起。但是**如果结果在主函数接收之前到达，则发送可能会失败，因为没有人准备好。**


这个问题是所谓的竞争条件的教科书示例，但解决方法很简单。我们只需要确保缓冲通道ch（通过将缓冲区长度添加为make的第二个参数），保证第一次发送有地方放置该值。这确保发送始终成功，并且无论执行顺序如何，都将检索第一个到达的值。  

这两个例子展示了Go可以简单表达GOroutine之间的复杂交互。  

- 原文地址：https://go.dev/blog/concurrency-timeouts
