### Go高级并发模式：第三部分（通道）  

今天我将身图探讨Go通道哈select声明的表现力。为了演示仅使用这两个原语可以实现多少，我将从头开始重写包sync。  

- 在这样做是时，我只是要接受一些妥协：  
	- 这篇文章是关于表现力的，而不是速度。这些实例将这却表达原理，但可能不如真实事例快。  
	- 与切片不同，通道没有固定大小的等效类型。虽然您可以在GO中有一个类型，但不能有一个大小chan int为4的类型，所以我不会有有效的零值，所有[4]int类型都需要构造函数调用。有一个案例可以解决这个问题，但这超出了本文的范围。  
	- 大多数类型只是通道，但没有什么能阻止您将这些通道隐藏在不透明的结构中以避免误用。由于这是一篇博文，而不是一个实际的库，为了简洁明了，我将使用裸通道。  
	- 我不会解释原语应该做什么，但我会连接到它们的官网文档。  

让我们开始吧！  

#### Once
Once是一个相当简单的原语：第一次调用Do(func())将导致所有其他并发调用阻塞，知道Do返回的参数。发生这种情况后，所有被阻止的呼叫和连续的呼叫将不执行任何操作并立即返回。  

这对于延迟初始化和单例实例游泳。  

让我们看一下代码：  
```go
type Once chan struct{}

func NewOnce() Once {
	o := make(Once, 1)
	o <- struct{}{}
	return o
}

func (o Once) Do(f func()) {
	_, ok := <-o
	if !ok {
		return
	}

	f()

	close(o)
}
```  

#### Mutex互斥体  

对于Mutex，我想说两点题外话：  
  - 大小为N的信号量最多允许N个goroutine在任何给定的时间保持其锁定。互斥体是大小为1的信号量的特殊情况。  
  - 互斥锁可能受益于TryLock方法。  

sync包不提供信号量或triable锁，但由于我们正在尝试证明通道的表达能力，因此select让我们实现两者。  
```go

type Semaphore chan struct{}

func NewSemaphore(size int) Semaphore {
	return make(Semaphore, size)
}

func (s Semaphore) Lock() {
	// Writes will only succeed if there is room in s.
	s <- struct{}{}
}

// TryLock is like Lock but it immediately returns whether it was able
// to lock or not without waiting.
func (s Semaphore) TryLock() bool {
	// Select with default case: if no cases ar ready
	// just fall in the default block
	select {
	case s <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s Semaphore) Unlock() {
	// Make room for other users of the semaphore
	<-s
}

```

如果我们现在想基于此实现互斥锁，我们可以执行一下操作：  
```go
type Mutex Semaphore

func NewMutex() Mutex{
	return Mutex(NewSemaphore(1))
}
```  
#### **读写互斥锁**  

RWMutex是一个稍微复杂的原语：它允许任意数量多并发读锁，但在任何给定时间只能有一个写锁。还可以保证，如果有人持有写锁，任何人都不应该能够拥有货获取读锁。  

标准库实现还允许，如果尝试写锁，进一步的读锁将排队并等待以避免写锁匮乏。为简洁起见，我们将放宽次条件。有一种方法可以通过使用1大小的通道作为互斥锁来保护RWLock状态，但这是一个无聊的例子。如果总是至少有一个读取器，则此实现将使编写起挨饿。  

注意：此代码的灵感来自于Bryan Mills对此概念的实现。  

RWMutex有三种状态：免费，有写和有读。这意味着我们需要两个通道：当互斥锁空闲时，我们将两个通道都为空；当有一个writer时，我们将有一个带有孔结构的通道；当有读时，我们将在两个通道中都有一个值，其中一个是读计数。  

```go

type RWMutex struct {
	write   chan struct{}
	readers chan int
}

func NewLock() RWMutex {
	return RWMutex{
		// This is used as a normal mutex.
		write:   make(chan struct{}, 1),
		// This is used to protect the readers count. 
		// By receiving the value it is guaranteed that no
		// other goroutine is changing it at the same time.
		readers: make(chan int, 1),
	}
}

func (l RWMutex) Lock() {
	l.write <- struct{}{}
}

func (l RWMutex) Unlock() {
	<- l.write
}

func (l RWMutex) RLock(){
	// Count current readers. Default to 0.
	var rs int
	// Select on the channels without default.
	// One and only one case will be selected and this
	// will block until one case becomes available.
	select {
	case l.write <- struct{}{}: // One sending case for write.
		// If the write lock is available we have no readers.
		// we grab the write lock to prevent concurrent
		// read-writes.
	case rs = <-l.readers:
		//	There already ar readers, let's grab and update the
		// readers count.
		
	}
	
	rs++
	
	l.readers <- rs
}

func (l RWMutex) RUnlock(){
	rs := <- l.readers
	rs--
	if rs == 0 {
		<- l.write
		return
	}
	
	l.readers <- rs
}
```


TryLock 方法可以TryRLock向我们在前面示例中所做的那样，通过将默认情况添加到通道操作来实现：  

```go

func (l RWMutex) RUnlock() {
	rs := <-l.readers
	rs--
	if rs == 0 {
		<-l.write
		return
	}

	l.readers <- rs
}

func (l RWMutex) TryLock() bool {
	select {
	case l.write <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l RWMutex) TryRLock() bool {
	var rs int
	select {
	case l.write <- struct{}{}:
	case rs = <-l.readers:
	default:
		return false

	}
	rs++
	l.readers <- rs
	return true
}


```  

#### POOL  

[池](https://pkg.go.dev/sync#Pool)用于减轻垃圾收集器的压力，并重用经常分配和销毁的对象。  

为此，标准库使用了许多技术：线程本地存储；删除过大的对象和生成，使对象仅在两个垃圾回收之间使用时才在池中存活。  

我们不会提供所有这些功能：我们无权访问线程本地存储或垃圾回收信息，也不关心。我们的池将是固定大小的，因为我们只想表达类型的语义，而不是实现细节。  

我们将提供Pool没有的一个实用程序是清理功能。  

当使用Pool不清除从中产生的对象时，这种情况很常见，从而导致非零内存的重用，这可能会导致令人讨厌的错误和漏洞。我们的视线将保证当且仅当返回的对象被回收时调用更干净的函数。  

例如，这可以用于切片以将切片重新切片为0长度，或用于结构以将所有字段归零。  

如果位置定清理，我们将不调用它。  

```go
package main

import "fmt"

type Item = interface{}

type Pool struct {
	buf   chan Item
	alloc func() Item
	clean func(Item) Item
}

func NewPool(size int, alloc func() Item, clean func(Item) Item) *Pool {
	return &Pool{
		buf:   make(chan Item, size),
		alloc: alloc,
		clean: clean,
	}
}

func (p *Pool) Get() Item {
	select {
	case i := <-p.buf:
		if p.clean != nil {
			return p.clean(i)
		}
		return i
	default:
		return p.alloc()
	}
}

func (p *Pool) Put(x Item) {
	select {
	case p.buf <- x:
	default:

	}
}

``` 
一个使用示例是： 
```go

func test() {
	p := NewPool(1024,
		func() interface{} {
			return make([]byte, 0, 10)
		},
		func(i interface{}) interface{} {
			return i.([]byte)[0]
		})
	fmt.Println(p)
}
``` 

#### MAP

对于这篇文章来说，[MAP](https://pkg.go.dev/sync#Map)不是很有趣，因为它在语义上等同于带有a的一个map RWMutex. 该map类型基本上是一个字典，针对预计读取次数多于写入次数的用例进行了优化。它只是一种使某些特定用例更快的技术，因此我们将跳过它，因为今天我们不关心速度。


#### WaitGroup

不知什么原因，[WaitGroup](https://pkg.go.dev/sync#WaitGroup)最终总是成为实现起来最复杂的基元之一。这也是发生在我的[Javascript's Atomic post](https://blogtitle.github.io/using-javascript-sharedarraybuffers-and-atomics/)帖子中，我在其中放置了一个竞争条件WaitGroup.  


等待组允许多种用途，但最常见的用途时创建一个组，对它进行计数，生成与该计数一样多的goroutine，Add然后等待所有goroutine完成。每次gorutine运行完毕，他都会调用Done该组来表示它已经完成了工作。等待组可以通过呼叫或以负计数（甚至大于-1）呼叫Done Add来达到0计数。当组达到0时，当前在Wait呼叫中被阻止的所有服务员都将恢复执行。将来的Wait调用不会被阻止。  

一个鲜为人知的功能是等待组可以重用：Add在计数器达到0后仍然可以调用，他会将等待组重新置于阻塞状态。

这意味着每个给定的等待组，我们都有一种“生成”：
  - 当计数器从0移动到正数是，一代就开始了
  - 当家暑期达到0时，一代就结束了
  - 当一代人结束时，那一代人的所有服务员都会被解封  

让我们看看代码：  
```go
type generation struct {
	// A barrier for waiters to wait on.
	// This will never be used for sending, only receive and close.
	wait chan struct{}
	// The counter for remaining jobs to wait for.
	n int
}

func newGeneration() generation {
	return generation{ wait: make(chan struct{}) }
}
func (g generation) end() {
	// The end of a generation is signalled by closing its channel.
	close(g.wait)
}

// Here we use a channel to protect the current generation.
// This is basically a mutex for the state of the WaitGroup.
type WaitGroup chan generation

func NewWaitGroup() WaitGroup {
	wg := make(WaitGroup, 1)
	g := newGeneration()
	// On a new waitgroup waits should just return, so
	// it behaves exactly as after a terminated generation.
	g.end()
	wg <- g
	return wg
}

func (wg WaitGroup) Add(delta int) {
	// Acquire the current generation.
	g := <-wg
	if g.n == 0 {
		// We were at 0, create the next generation.
		g = newGeneration()
	}
	g.n += delta
	if g.n < 0 {
		// This is the same behavior of the stdlib.
		panic("negative WaitGroup count")
	}
	if g.n == 0 {
		// We reached zero, signal waiters to return from Wait.
		g.end()
	}
	// Release the current generation.
	wg <- g
}

func (wg WaitGroup) Done() { wg.Add(-1) }

func (wg WaitGroup) Wait() {
	// Acquire the current generation.
	g := <-wg
	// Save a reference to the current waiting chan.
	wait := g.wait
	// Release the current generation.
	wg <- g
	// Wait for the chan to be closed.
	<-wait
}
```  

#### Condition 条件

[Cond](https://pkg.go.dev/sync#Cond) 是同步包中最具争议的类型。我认为这是一个危险的原语，它太容易错误地使用。我从不使用它，因为我不相信自己能正确使用它，在代码审查期间，我总是建议使用其他原语来代替。甚至 Bryan Mills（他是 Go 团队的一员，在原语方面 sync 做了很多工作）也提议删除它。


我想存在的最重要原因是 Cond 频道不能重新开放播放两次，但我不确定这种好处是否值得付出代价。  

即使不考虑这容易出错的事实，它也不适用于频道（关于这个问题的问题）。例如，目前无法选择 a Cond 和上下文取消：它需要一些包装和额外的 goroutines，这些都很昂贵并且可能会被泄露。  

这个 API 还有更多奇怪的地方：它要求其用户自己完成部分同步。引用文档：
> 每个 Cond 都有一个关联的 Locker L（通常是 \*Mutex 或 \*RWMutex），在更改条件和调用 Wait 方法时必须保留该 Locker L。   

这意味着我们不必关心用户更改 Cond L 领域或在通话中 Wait 比赛，但可能会有比赛 Broadcast ，我们将 Signal 解决这些问题。  


如果我们考虑以下文档 Wait ，情况会变得更糟：  
> 等待原子解锁 c.L 并暂停调用 goroutine 的执行。稍后恢复执行后，Wait 在返回之前锁定 c.L。由于 c.L 在 Wait 首次恢复时未被锁定，因此当 Wait 返回时，调用方通常不能假定条件为 true。相反，调用方应在循环中等待。  

为了实现这个奇怪的原语，我将使用通道的通道。  

```go
type Locker interface {
	Lock()
	Unlock()
}

type barrier chan struct{}

type Cond struct {
	L  Locker
	bs chan barrier
}

func NewCond(l Locker) Cond {
	c := Cond{
		L:  l,
		bs: make(chan barrier, 1),
	}
	// Waits will block until signalled to continue.
	c.bs <- make(barrier)
	return c
}

func (c Cond) Broadcast() {
	// Acquire barrier.
	b := <-c.bs
	// Release all waiters.
	close(b)
	// Create a new barrier for future calls.
	c.bs <- make(barrier)
}

func (c Cond) Signal() {
	// Acquire barrier.
	b := <-c.bs
	// Release one waiter if there are any waiting.
	select {
	case b <- struct{}{}:
	default:
	}
	// Release barrier.
	c.bs <- b
}

// According to the doc we have to perform two actions atomically:
// * Call Unlock
// * Suspend execution
// To do so we receive the current barrier, call Unlock while we still
// hold it and release it. This guarantees that nothing else has happened
// in the meantime.
// After this operation we wait on the barrier we received, which
// might not reflect the current one (as intended).
func (c Cond) Wait() {
	// Acquire barrier.
	b := <-c.bs
	// Unlock while in critical section.
	c.L.Unlock()
	// Release barrier.
	c.bs <- b
	// Wait for release on the value of barrier that was valid during
	// the call to Unlock.
	<-b
	// We were unblocked, acquire lock.
	c.L.Lock()
}
```  

#### Conclusion 结论

Go 通道和组合非常具有表现力，并允许以非常富有表现力的方式创建高级同步和 select 编排原语。我认为许多设计选择，比如具有有限大小的通道或可以同时接收和发送的选择，为Go内置类型提供了在其他语言中很少见的东西。


- 原文链接：https://blogtitle.github.io/go-advanced-concurrency-patterns-part-3-channels/


