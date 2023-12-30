## 并发模型之谁发起谁负责（Leave concurrency to the caller）

- 这两个API有什么区别？

```go
// ListDirectory returns the contents of dir.
func ListDirectory(dir string) ([]string, error)


// ListDirectory returns a channel over which
// 读取一条发送一条，当列表条目用尽，通道将被关闭
func ListDirectory(dir string) chan string

```
首先明显的差异，第一个示例将目录读取到切片中，然后返回整个切片，或者出现问题时返回错误。这是同步发生的，ListDirectory调用者直到目录中所有条目被读取，才会获得答案。**根据目录的大小，这可能需要很长时间，并且可能会分配大量的内存来构建目录文件的变量**  

让我们来看第二个例子。这有点像GO，返回一个通道，ListDirectory将条目将通过该通道传递。当通道关闭时，表示没有更多的目录条目。由于通道的填充发生在返回之前，Listdirectory因此可能启动一个goroutine来填充通道。  

> 第二个版本没有必要实际使用Goroutine；它可以分配一个足以容纳所拥有条目而不阻塞的通道，填充通道，关闭它，然后将通道返回给调用方。但这不太可能，也因为会消耗大量的内存来缓冲通道中所有结果，存在于同步返回相同的问题。  

- 频道版本ListDirectory还有两外两个问题：
	- **信号不明确：**通过使用关闭通道作为没有更多条目要处理的信号，ListDirectory无法通过通道告诉调用方返回集不完整，因为途中遇到错误。调用方无法分辨空目录和完全从目录中读取的错误之间的区别。两者都会导致返回一个通道，ListDirectory该通道似乎立即关闭
	- **关闭通道难：**调用方必须从通道继续读取，直到通道关闭，因为这是调用房知道让通道goroutine停止的唯一方法。这是对使用者的严重限制，及时呼叫者可能已经收到它想要的答案，也必须话事件从频道中读取。就大型目录占用内存而言，它可能更有效，但这种方法并不比原始的基于分片方法快。 

- 这两种问题的解决方案时使用回调，该函数在读取每个条目的上下文中使用
```go
func ListDirectory(dir string, fn func(string))
``` 
毫不奇怪，filepath.WalkDir这就是该函数的工作原理
> 如果函数启动goroutine，则必须为调用房提供显式停止该Goroutine的方法。将异步执行函数的决定权交给该函数的调用方通常更容易。  

#### 思考（见并发模型之开启和关闭）
- ListDirectory如何具体实现呢？
- 调用者在拿到想要的数据后，怎么主动关闭？

#### 参考文献：  
https://dave.cheney.net/practical-go/presentations/qcon-china.html#_concurrency
