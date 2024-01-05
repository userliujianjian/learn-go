### 介绍Go竞争条件探测器(race detector)

#### 简介
竞争条件是最阴险和最难以捕捉的变成错误之一。它们通常会导致不稳定和神秘的故障，通常是在代码部署到生产环境很久之后被发现。虽然go的并发机制使编写干净的并发代码变得更容易，但它们并不能阻止竞争条件。需要小心、勤奋和测试。工具可以提供帮助。  

很开心的宣布，Go1.1包括一个竞争检测器，这是一种用于在Go代码中查找竞争条件的新工具。它目前可用于具有64位X86处理器的Linux、os X和Windows系统。  

竞争检测器基于C/C++ ThreadSanitizer运行时库，该库已用于检测Google内部代码库和Chromium中的许多错误。该技术于2012年9月与Go集成；从那时起，它已经在标准库检测到了42个竞争条件。它现在是我们持续构建过程的一部分，它会继续在出现竞争条件时发挥关键作用。  

#### 它如何工作
竞争检测器与Go工具链集成在一起。设置-race命令行标志后，编译器会使用记录访问内存的时间和方式的代码来检测所有内存访问，而运行时库监视对共享变量的不同步访问。当检测到这种“不雅”行为时，会打印警告。[有关算法的详细信息，请参阅文本](https://github.com/google/sanitizers/wiki/ThreadSanitizerAlgorithm).  

由于其设计，竞争检测器只有在实际由运行代码触发时才能检测争用条件，这意味着在实际工作负载下运行启用争用的二进制文件非常重要。但是其勇争用的二进制文件可以使用十倍的CPU和内存，因此始终启用竞争检测器是不切实际的。摆脱这种困境的一种方法是在启用竞争检测器的情况下运行一些测试。负载测试和集成测试是很好的候选者，因为它们倾向于执行代码并发部分。使用上产工作负载的另一种方法是在运行的服务器池中部署单个启用竞争检测器的实例。  

#### **使用竞争检测器**  
竞争检测器与Go工具链完全集成。若要在启用竞争检测器的情况下生成代码，只需要将-race标志添加到命令行即可：  
```bash
go test -race mypkg
go run -race mysrc.go
go build -race mycmd
go install -race mypkg
```
要亲自试用竞争检测器，将此示例程序肤质到racy.go:  
```go
package main

import "fmt"

func main() {
	done := make(chan bool)
	m := make(map[string]string)
	m["name"] = "world"
	go func() {
		m["name"] = "data race"
		done <- true
	}()

	//<-done
	fmt.Println("Hello, ", m["name"])
	// data race 的原因是因为goroutine 跟主goroutine同时写map
	<-done // 无缓冲通道，当接收准备好之后，发送才开始执行。 
}
```

```bash
go run -race racy.go
```

output:  
```text
Hello,  world
==================
WARNING: DATA RACE
Write at 0x00c00007c0c0 by goroutine 6:
  runtime.mapassign_faststr()
      /usr/local/opt/go/libexec/src/runtime/map_faststr.go:203 +0x0
  main.main.func1()
      /golang/example/routine/memory/race_detector.go:10 +0x4a

Previous read at 0x00c00007c0c0 by main goroutine:
  runtime.mapaccess1_faststr()
      /usr/local/opt/go/libexec/src/runtime/map_faststr.go:13 +0x0
  main.main()
      /golang/example/routine/memory/race_detector.go:14 +0x159

Goroutine 6 (running) created at:
  main.main()
      /golang/example/routine/memory/race_detector.go:9 +0x13c
==================
==================
WARNING: DATA RACE
Write at 0x00c000108088 by goroutine 6:
  main.main.func1()
      /golang/example/routine/memory/race_detector.go:10 +0x56

Previous read at 0x00c000108088 by main goroutine:
  main.main()
      /golang/example/routine/memory/race_detector.go:14 +0x164

Goroutine 6 (running) created at:
  main.main()
      /golang/example/routine/memory/race_detector.go:9 +0x13c
==================
Found 2 data race(s)
exit status 66
```

## 示例(Example):  
以下是竞争检测器捕捉到两个实际问题的例子。 

#### **示例1: Timer.Reset**  
第一个示例是竞争检测器发现实际错误的简化版本。它使用计时器在0到1秒之间的随机持续时间后打印消息。它重复这样做五秒钟。它用于time.AfterFunc Timer位第一条消息创建一个，然后使用该Reset方法安排下一条消息，每次都重用。 Timer  
```go
func main() {
	start := time.Now()
	var t *time.Timer
	t = time.AfterFunc(randomDuration(), func() {
		fmt.Println(time.Now().Sub(start))
		t.Reset(randomDuration())
	})
	time.Sleep(5 * time.Second)
}

func randomDuration() time.Duration {
	return time.Duration(rand.Int63n(1e9))
}

```
这看起来像是合理的代码，但在某些情况下，它以一种令人惊讶的方式失败了： 

```text
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xb code=0x1 addr=0x8 pc=0x41e38a]

goroutine 4 [running]:
time.stopTimer(0x8, 0x12fe6b35d9472d96)
    src/pkg/runtime/ztime_linux_amd64.c:35 +0x25
time.(*Timer).Reset(0x0, 0x4e5904f, 0x1)
    src/pkg/time/sleep.go:81 +0x42
main.func·001()
    race.go:14 +0xe3
created by time.goFunc
    src/pkg/time/sleep.go:122 +0x48
```
这是怎么回事？在启用竞争检测器的情况下运行程序更具启发性：  
```text
==================
WARNING: DATA RACE
Read at 0x00c000092030 by goroutine 7:
  main.example1.func1()
      /golang/example/routine/memory/race_detector.go:34 +0xbe

Previous write at 0x00c000092030 by main goroutine:
  main.example1()
      /golang/example/routine/memory/race_detector.go:32 +0x15c
  main.main()
      /golang/example/routine/memory/race_detector.go:11 +0x1c

Goroutine 7 (running) created at:
  time.goFunc()
      /usr/local/opt/go/libexec/src/time/sleep.go:176 +0x44

```

竞争检测器显示了问题，来自不同的goroutine的变量t的读取和写入不同步。如果出事计时器持续的时间非常小，则计时器函数可能会在主goroutine赋值t之前出发，因此调用t.Reset时为nil 。  

为了解决竞争条件，我们将代码更改为仅从主goroutine读取和写入变量t：  
```go
func example_1_solution() {
	start := time.Now()
	reset := make(chan bool)
	var t *time.Timer

	t = time.AfterFunc(randomDuration(), func() {
		fmt.Println(time.Now().Sub(start))
		reset <- true
	})

	for time.Since(start) < time.Second*5 {
		<-reset
		t.Reset(randomDuration())
	}
}
```
在这里，主要的goroutine全权负责设置和重置，Timer t新的重置通道传达了以线程安全方式重制计时器的需要。  

- **一种更简单但有效的方法是避免重复使用计时器**
#### **示例2: ioutil.Discard**
第二个示例更微妙。  
ioutil包Discard的对象实现io.Writer,但丢弃写入它的所有数据。 可以把它想象成/dev/null: 一个发送你需要读取但不想存储的数据的地方。它通常用于io.Copy排空阅读器，如下所示：  
```go
io.Copy(ioutil.Discard, reader)
```

早在2011年7月，Go团队就注意到以这种方式使用Discard效率低下： 该Copy函数每次调用时都会分配一个内存32KB缓冲区，但当Discard缓冲期一起使用时，这是不必要的，因为我们只是丢弃了读取的数据。我们认为这种管用的Copy用Discard法不应该如此昂贵。  

解决方法很简单。如果给定Writer实现了一个ReadFrom方法，则Copy调用如下：  
```go
io.Copy(writer, reader)
```
委托给这个可能更有效的调用：  
```go
writer.ReadFrom(read)
```

我们在向Discard的基础类型添加了一个ReadFrom方法，该方法具有一个在所有用户之间共享的内部缓冲区。我们知道这是在理论上一个竞争条件，但由于所有对穿冲区的写入都应该被丢弃，我们认为这并不重要。  
当竞争检测器被实施时，它立即将此代码标记为racy。同样我们认为代码可能有问题，但认为竞争条件不是“真实的”。为了避免构建中的“误报”，我们实现了一个非活泼版本，该版本仅在竞争检测器运行时启用。  

但是几个月后，布拉德遇到了一个令人沮丧和奇怪的错误。经过几天的测试，他将其缩小到由ioutil.Discard。 

这是已知代码io/ioutil,其中Discard是一个devNull在所有用户之间共享一个缓冲区。  
```go
var blackHole [4096]byte // shared buffer

func (devNull)ReadFrom(r io.Reader)(n int64, err error){
	readSize := 0
	for {
		readSize, err = r.Read(blackHole[:])
		n += int64(readSize)
	}
	if err != nil {
		if err == io.EOF{
			return n, nil
		}
		return 
	}
	return 
}
```
Brad的程序包括一个trackDigestReader类型，它包装并io.Reader记录它所读取内容的哈希摘要。  
```go
type trackDigestReader struct{
	r io.Reader
	h hash.Hash
}

func(t trackDigestReader) Read(p []byte)(n int, err error){
	n, err = t.r.Read(p)
	t.h.Write(p[:n])
	return
}
```  

例如，它可用于在读取文件时计算文件的SHA-1哈希值

```go
tdr := trackDigestReader{r: file, h: sha1.New()}
io.Copy(writer, tdr)
fmt.Printf("File hash: %x", tdr.h.Sum(nil))
```  

在某些情况下，将无处写入数据，但仍然要对文件进行哈希处理，因此Discard将使用

```go
io.Copy(ioutil.Discard, tdr)
```  

在这种情况下，blackHole缓冲区不仅仅是一个黑洞；它是从源io.Reader读取数据到将其写入hash.Hash当多个goroutine同时对文件进行哈希处理时，每个文件共享相同的blackHole缓冲区，竞争条件通过破坏读取和哈希之间的数据来表现自己。 没有发生错误或者恐慌，但是哈希值是错误的。讨厌！  



```go
func (t trackDigestReader) Read(p []byte) (n int, err error) {
    // the buffer p is blackHole
    n, err = t.r.Read(p)
    // p may be corrupted by another goroutine here,
    // between the Read above and the Write below
    t.h.Write(p[:n])
    return
}
```  

该错误最终通过为每次使用 提供唯一的缓冲区来修复 ioutil.Discard ，消除了共享缓冲区上的争用条件。


### **结论**
竞争检测器是检查并发程序正确性的强大工具。它不会发出误报，因此请认真对待其警告。但它只和你的测试一样好;您必须确保它们彻底执行代码的并发属性，以便争用检测器可以完成其工作。

