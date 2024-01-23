### Go并发模式：管道和取消

#### 介绍
Go的并发原语使构建流数据管道变得容易，这些管道可以有效地利用I/O和多个CPU。本文介绍了此类管道的示例，重点介绍了操作失败时出现的微妙之处，并介绍了干净地处理故障的结束。  

#### 什么是管道？
Go中没有流水线的正式定义；它只是众多并发程序中的一种。非正式地，流水线是由通道连接的一系列阶段，其中每个阶段都是一组运行相同函数的goroutine。在每个阶段，goroutines
  - 通过入站通道从上游接收值  
  - 对该数据执行某些功能，通常会产生新值  
  - 通过出战通道向下游发送值  

 每个阶段都有任意数量多入站和出站通道，但第一和最后一个阶段除外，它们分别只有出站和入站通道。第一阶段有时称为源或生产者；最后一个阶段，接收器或消费者。  

 我们将从一个简单的示例管道开始，以解释这些想法和技术。稍后，我们将展示一个更现实的例子。  

 #### 平方数  
 考虑具有三个阶段的通道    
 第一阶段gen是一个将整数列表转换为发出列表中整数的通道函数。该gen函数启动一个goroutine，在通道上发送整数，并在发送完成后关闭通道：  

 ```go
func gen(nums ...int) <- chan int {
	out := make(chan int)
	go func(){
		for _, n := range nums{
			out <- n
		}
		close(out)
	}()

	return out
}
 ```  


 第二阶段，sq从通道接收整数并返回一个发出每个接收到的整数的平方的通道。在入站通道关闭且该阶段已向下游发送所有值后，它会关闭出站通道：  
```go
func sq(in <- chan int) <- chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()

	return out
}
 ```  

 该main函数设置管道并运行最后阶段：它从第二阶段接收值并打印每个值，知道通道关闭：  

 ```go
func main() {
	c := gen(2, 3)
	out := sq(c)

	fmt.Println(<-out) // output: 4
	fmt.Println(<-out) // output: 9

}
 ```  


 优于sq其入站和出站通道具有相同的类型，因此我们可以组合它任意多次。我们还可以重写main为范围循环，就像其他阶段一样：  

 ```go
func main() {
	for n := range sq(sq(gen(2, 3))) {
		fmt.Println(n) // output: 16、81
	}
}
 ```  


 #### 扇出、扇入
*多个函数可以从同一通道读取，直到该通道关闭*； 这叫**扇出**。这提供了一种在一组工作线程之间分配工作的方法，以并行化CPU使用和IO。  

函数可以从多个输入中读取数据，并通过将输入通道复用到单个通道上来继续执行，直到所有输入都关闭，该通道所在所有输入都关闭时关闭。称为**扇入**。

我们可以将管道更改为运行两个示例sq，每个实例都从同一输入通道读取。我们引入了一个新函数merge来分散结果：  
```go
func main() {
	in := gen(2, 3)
	c1 := sq(in)
	c2 := sq(in)

	for n := range merge(c1, c2) {
		fmt.Println(n)
	}
}
```  
该merge函数通过每个入站通道启动一个goroutine，将值复制到为一到出站通道，将通道列表转换为单个通道。启动所有output goroutine后，在该通道上完成所有发送后，merge在启动一个goroutine以关闭出站通道。  

在关闭通道上发送panic，因此请务必确保在调用close之前完成所有发送。该sync.WaitGroup类型提供了一种安排此同步到简单方法：  
```go 
func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	output := func(c <-chan int) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))

	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
```

#### **短暂停止**  
我们的管道函数有一个模式：  
  - 当所有发送操作完成后，阶段关闭其出站通道。
  - 阶段不断从入站通道接收值，知道这些通道关闭。  

这种模式允许每个接收阶段编写为循环range，并确保一旦所有值成功发送到下游，所有goroutine都会退出。  

但在实际管道中，阶段并不总是接收所有入站值。有时这是设计使然：接受者可能只需要值的子集即可获得进展。更常见的是，阶段提前退出，因为入站值表示较早阶段中的错误。在任何一种情况下，接收器都不必等待剩余值到达，并且我们希望早期阶段停止生成后期阶段不需要的值。  

在我们的示例管道中，如果某个阶段无法消耗所有入站值，则尝试发送这些值的goroutine将无限期阻塞：  
```go
out := merge(c1, c2)
fmt.Println(<- out)
return
```  
这是资源泄漏：goroutine会消耗内存和运行时资源，而goroutine堆栈中的堆饮用会组织数据被垃圾回收。Goroutines不是垃圾回收；她们必须自行退出。  

我们需要安排管道的上游阶段退出，即使下游阶段无法接受到所有入站值。一种方法是将出站通道改为具有缓冲区。  
缓冲区可以保存固定数量的值；如果缓冲区中有空间，则发送操作会立即完成：  
```go
c := make(chan int, 2)
c <- 1
c <- 2
c <- 3 // 3会一直被阻塞，知道有其他goroutine接收到1
```  

在通道创建时知道要发送的值的数量时，缓冲区可以简化代码。例如，我们可以重写gen以将整数列表复制到缓冲通道中，并避免创建新的goroutine：  
```go
func gen(nums ...int) <- chan int{
	out := make(chan int, len(nums))
	for _, n := range nums{
		out <- n
	}

	close(out)
	return out
}

```  
> 有缓冲通道特征：发送在接收完成之前完成。  

回到管道中被阻塞的goroutines，我们可以考虑像outbound通道添加一个缓冲区，由merge：  
```go
func merge(cs ...<-chan int) <- chan int{
	var wg sync.WaitGroup
	out := make(chan int, 1)
	// ...
}
```  
这虽然修复了该程序中被阻塞的goroutine，但这是糟糕的代码。此处缓冲区大小的选择1取决于了解将接受的值数以及下游阶段将消耗的值merge数。这是脆弱的：如果我们将额外的值传给gen，或者如果下游阶段读取的值更少，我们将再次阻塞goroutines。  

相反，我们需要为下游阶段提供一种方法，以向发送着表明它们将停止接受输入。  

#### 明确取消  

当main决定不接受所有值当情况下退出out时，它必须告诉上游阶段的goroutine放弃他们试图发送的值。它们通过在名为done的通道上发送值来实现。它发送两个值，因为可能有两个被阻塞的发件人：  

```go
func main(){
	in := gen(2, 3)
	c1 := sq(in)
	c2 := sq(in)

	done := make(chan struct{}, 2)
	out := merge(done, c1, c2)
	fmt.Println(<-out)

	done <- struct{}{}
	done <- struct{}{}
}

```  
发送goroutine将其发送操作替换为一个select语句，该语句发送发生在out时活从done接收到值时继续执行。  
类型done是空结构，因为该值无关紧要：它是指示应放弃发送的out接收事件。output goroutines继续在其入站通道C上循环，因此上游阶段不会被阻塞。（我们稍后讨论如何让这个循环提前返回）：  

```go
func merge(done <-chan struct{}, cs ...<-chan int) <- chan int{
	var wg sync.Waitgroup
	out := make(chan  int)
	output := func(c <- chan int) {
		for n := range c {
			select {
			case out <- n:
			case <- done:
			}
		}
		wg.Done()
	}
	// ....
}

```  
这种方法有一个问题：每个下游接收方都需要知道可能被组织的上行发送方的数量，并安排在提前返回时向这些发送方发出信号。跟踪这些计数计繁琐又容易出错。  

我们需要一种方法来告诉未知且无限数量的goroutine停止向下游发送它们的值。在go中，我们可以通过关闭通道来做到这一点，**因为关闭通道上的接受操作总是可以立即进行，从而产生元素类型的零值。**  

这意味着main只需关闭done通道即可解锁所有发件人。这种关闭实际上是对发送着的广播信号。我们将每个流水线函数扩展为接受done作为参数，并通过defer语句安排关闭发生，一边所有返回路径都将main向流水线阶段发出退出信号。  

```go
func main(){
	done := make(chan struct{})
	defer close(done)

	in := gen(done, 2, 3)

	c1 := sq(done, in)
	c2 := sq(done, in)

	out := merge(done , c1, c2)
	fmt.Println(<-out)

	// done will be closed by the deferred call.
}

```  

现在，我们的每个管道阶段都可以在关闭后done立即免费返回。output merge历程中可以在不消耗其入站通道的情况下返回，因为他知道上游发送方，将在sq关闭时done停止尝试发送。output ensure wg.Done通过语句defer在所有返回路径上调用：  

```go
func merge(done <-chan struct{}, cs ...<-chan int) <-chan int{
	var wg sync.WaitGroup()
	out := make(chan int)

	output := func(c <- chan int){
		defer wg.Done()
		for n := range c {
            select {
            case out <- n:
            case <-done:
                return
            }
          }
        }
    // 
}

```  
同样，sq可以在关闭后done立即返回。sq通过defer语句确保其out通道在所有返回路径上关闭：  


```go
func sq(done <-chan struct{}, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            select {
            case out <- n * n:
            case <-done:
                return
            }
        }
    }()
    return out
}

```  
- **以下是管道建设的准则**
	- 完成所有发送操作后，阶段将关闭其出站通道。
	- 阶段不断接受来自入站通道的值，知道这些哦你管道关闭或发送方被解除阻止。  

管道通过确保**有足够的缓冲区**来发送所有发送的值，或者通过在接收方可能**放弃通道时显式向发送方发出信号**来取消阻止发送方。

#### **消化一棵树**
让我们考虑一个更现实的通道。  

MD5是一种消息摘要算法，可用作文件校验。命令行使用程序md5sum打印渐渐列表的摘要值。  

```bash
% md5sum *.go
d47c2bbc28298ca9befdfbc5d3aa4e65  bounded.go
ee869afd31f83cbb2d10ee81b2b831dc  parallel.go
b88175e65fdcbc01ac08aaf1fd9b5e96  serial.go
```  
我们的示例程序类似于md5sum，而且将单个目录作为参数，并打印该目录下每个常规文件的摘要值，并按路径名排序。  

```bash
% go run serial.go .
d47c2bbc28298ca9befdfbc5d3aa4e65  bounded.go
ee869afd31f83cbb2d10ee81b2b831dc  parallel.go
b88175e65fdcbc01ac08aaf1fd9b5e96  serial.go
```  
我们程序的main函数调用一个helper函数，该函数MD5ALL返回从路径名摘要值的映射，然后对结果进行排序并打印：  

```go
func main() {
	m, err := MD5ALL(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	var paths []string
	for path := range m {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	for _, path := range paths {
		fmt.Printf("%x %s \n", m[path], path)
	}
}

```  
MD5ALL功能是我们讨论的重点。在serial.go中，该实现不使用并发，只是在便利树时读取和求和每个文件。  

```go
func MD5ALL(root string) (map[string][md5.Size]byte, error) {
	m := make(map[string][md5.Size]byte)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		m[path] = md5.Sum(data)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return m, nil

}
```  
#### **平行消解**  
在paraller.go， 我们拆分MD5ALL为两个阶段的管道。第一阶段，遍历树，sumFiles在新goroutine中笑话每个文件，并将结果发送到值类型为：result

```go
type result struct{
	path string
	sum [md5.Size]byte
	err error
}
```  
sumFiles返回两个通道：一个用于返回，results另一个用于filepath.walk返回的错误。walk函数启动一个新的goroutine来处理每个常规文件，然后检查done 如果done关闭，步骤将立即停止：  
```go
func sumFiles(done <-chan struct{}, root string) (<-chan result, <-chan error) {
	c := make(chan result)
	errc := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			wg.Add(1)
			go func() {
				data, err := os.ReadFile(path)
				select {
				case c <- result{path: path, sum: md5.Sum(data), err: err}:
				case <-done:
				}
				wg.Done()
			}()

			select {
			case <-done:
				return errors.New("walk canceled")
			default:
				return nil
			}

		})

		go func() {
			wg.Wait()
			close(c)
		}()

		errc <- err

	}()

	return c, errc
}
```  
MD5ALL接受来自c的照耀值。MD5ALL在错误的早起返回，done通过defer
```go

func md5all(root string) (map[string][md5.Size]byte, error) {
	done := make(chan struct{})
	defer close(done)

	c, errc := sumFiles(done, root)
	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}

	if err := <-errc; err != nil {
		return nil, err
	}

	return m, nil
}
```  

中间阶段启动固定数量的goroutines，这些digester goroutines从paths通道接收文件名饼子啊通道c上发送results：  

```go
func digester(done <-chan struct{}, paths <-chan string, c chan<- result) {
	for path := range paths {
		data, err := os.ReadFile(path)
		select {
		case c <- result{path: path, sum: md5.Sum(data), err: err}:
		case <-done:
			return
		}
	}
}
```  
与我们之前的示例不同，digester它不会关闭其输出通道，因为多个goroutines正在共享通道上发送。取而代之的是代码nil，md5all安排在完成所有操作digesters后关闭通道：  

```go
// Start a fixed number of goroutines to read and digest files.
    c := make(chan result)
    var wg sync.WaitGroup
    const numDigesters = 20
    wg.Add(numDigesters)
    for i := 0; i < numDigesters; i++ {
        go func() {
            digester(done, paths, c)
            wg.Done()
        }()
    }
    go func() {
        wg.Wait()
        close(c)
    }()
```  
相反，我们可以让每个消化器创建并返回自己的输出通道，但这样我们就需要额外的 goroutines 来扇入结果。  

最后阶段接收所有 来自 c ， results 然后检查来自 errc 的错误。此检查不能更早进行，因为在此之前， walkFiles 可能会阻止向下游发送值：  

```go
m := make(map[string][md5.Size]byte)
for r := range c {
    if r.err != nil {
        return nil, r.err
    }
    m[r.path] = r.sum
}
// Check whether the Walk failed.
if err := <-errc; err != nil {
    return nil, err
}
return m, nil

```

### 结论
本文介绍了在GO中构建刘数据管道技术。处理此类管道中的故障是很棘手的，因为管道中的每个阶段都可能组织尝试向下游发送值，并且下游阶段可能不再关心传入的数据。我们展示了关闭通道如何向流水线启动的所有goroutine广播完成信号，并定义了正确构建流水线的规则。  

- 扩展阅读
	- [GO并发模式](https://go.dev/talks/2012/concurrency.slide#1)  [视频](https://www.youtube.com/watch?v=f6kdp27TYZs)介绍了GO并发原语的基础知识以及应用他们的集中方法。  
	- [Advanced Go Concurrency Patterns 视频](https://go.dev/blog/io2013-talk-concurrency), 涵盖了GO原语复杂用法，尤其是select。  
	- Douglas McIlroy 的论文 [Squinting at Power Series](https://swtch.com/~rsc/thread/squint.pdf) 展示了类 Go 并发如何为复杂的计算提供优雅的支持


- 原文链接：https://go.dev/blog/pipelines





