### Go高级并发模式

Go并发模式三要素：
- 1. 谁调用谁负责开启
- 2. 什么时候会被关闭，如何结束
- 3. 是否有竞争

我们先来看一个示例：
- 清单1（ping pong）：
```go
type Ball struct {
	hits int
}

func main() {
	table := make(chan *Ball)
	go player("ping", table)
	go player("pong", table)

	table <- new(Ball)
	time.Sleep(1 * time.Second)
	<-table
	panic("show me the stacks")
}

func player(name string, table chan *Ball) {
	for {
		ball := <-table
		ball.hits++
		fmt.Println(name, ball.hits)
		time.Sleep(time.Millisecond * 100)
		table <- ball
	}
}

```
从清单1中可以看出，go开启一个goroutine是非常容易的， 上述代码也没有数据竞争，可以正常运行。
**大家有没有发现，goroutine的for循环何时退出呢？**

让我们来看看panic打印情况：  
```bash
...
pong 9
ping 10
pong 11
panic: show me the stacks

goroutine 1 [running]:
main.main()
        /Users/slice/GolandProjects/learn-go/golang/example/acvconc/pingpong/main.go:20 +0xd1

```
有goroutine正在运行，这是为什么呢？   
- 我们来看看player函数，函数中有一个for循环，导致panic主程序退出后goroutine依然一直在系统中运行

#### **总结：**    
- 开启goroutine非常容易，但怎么停下来呢？
- 长期存在的程序需要清理。  
- 让我们看看如何编写处理通信、周期性时间和取消的程序。  
- 核心是Go的select语句：就像一个switch，但决策是基于沟通能力做出的。 
```go
select{
	case xc <- x:
		// sent x on xc
	case y := <-yc:
		// received y from yc

}
```  







#### 让我们来看
