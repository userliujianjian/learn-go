## 并发陷阱2：不完整的工作

### 介绍
<u>介绍关于“不完整工作”的新陷阱。当程序在未完成的Goroutine（非主Goroutine）完成之前就终止是，就会发生不完整的工作。根据被强制终止的Goroutine的性质，这可能是一个严重的问题</u>

#### 未完成的工作
要查看未完成工作的简单示例，请查看此程序

- **清单1**
```go
func main(){
	fmt.Println("Hello")
	go fmt.Println("Goodbye")
}
```
清单1中，打印Hello之后，新开一个Goroutine打印GoodBye，程序立即达到函数末尾main并终止。如果你运行这个程序，你将不会看到Goodbye打印，因为Go规范中有一条规则：
> <u>程序执行首先初始化主包，然后调用函数main。当函数调用返回时，程序退出。它不会等待其他（非主）Goroutine完成</u>。

规范明确规定，当程序从函数返回时，您的程序不会等待任何未完成的goroutines完成main。这是一件好事！考虑一下让Goroutine泄漏或让Goroutine运行很长时间是多么容易。如果你的程序等待非主Goroutines完成才可以终止，那么它可能会陷入某种将是状态永远不会终止。

然而，当你启动一个Goroutine去做一些重要的事情，但该main函数不知道要等待它完成时，这种终止行为就会成为一个问题。此类情况可能会导致完整性问题，例如数据库、文件系统损坏或数据丢失。 

#### 一个真实的例子
在项目中，团队因为要跟踪某些事件构建了一个Web服务。记录时间的系统具有类似Tracker如清单2所示类型的方法
- **清单2**
```go
type Tracker stuct{}

func (t *Tracker) Event(data string){
	time.Sleep(time.Millisecond)
	log.Println(data)
}

```
客户担心跟中这些时间会增加相应时间和不必要的延迟，因此希望异步执行跟踪。**对性能作出假设是不明智的**，因此我们的首要任务就是通过以直接和同步的方法跟踪时间来测量服务的延迟。在这种情况下咽齿高的令人无法接受，团队决定需要采用异步方法。如果同步方法足够快，那么这个故事就会结束，因为我们会继续处理更重要的事情。

考虑到这一点，跟踪时间的处理程序最初是怎样编写的：

- **清单3**
```go

type App struct {
	track Tracker
}

func(a *App) Handle(w http.ResponseWriter, r *http.Request){

	w.WriteHeader(http.StatusCreated)

	go a.track.Event("this event")
}

```
清单3中代码的重要部分是第33航。这是a.track.Event在新Goroutine范围内调用该方法的地方。浙大到了异步跟踪时间的预期效果，而不会增加请求的延迟。然而，这段代码陷入了不完整工作的陷阱，必须重构。创建的任何Goroutine都不能保证运行或完成。这是一个完整性问题，因为服务器关闭，事件可能丢失

#### 重构保证
为了避免陷入困境，团队修改了类型Tracker来管理Goroutines本身。该类型使用了一个sync.WaitGroup 来记录打开的goroutine的技术，并提供一个ShutDown方法供main函数调用，该方法会等待所有Goroutines完成  

首先，处理程序被修改为不志杰创建goroutines，清单4中唯一的变化，就是不包含关键字go


- **清单4**
```go
func(a *App) Handle(w https.ResponseWriter, r *http.Request){

	w.WriteHeader(http.StatusCreated)
	a.track.Event("this event")

}

```
接下来，该Tracker类型被重写以管理Goroutines本身
- **清单5**
```go
type Tracker struct {
	wg sync.WaitGroup
}

func(t *Tracker) Event(data string) {
	t.wg.Add(1)

	go func() {
		defer t.wg.Done()

		time.Sleep(time.Millisecond)
		log.Println(data)
	}
}

func(t *Tracker) ShutDown(){
	t.wg.Wait()
}

```
清单5中，添加了sync.WaitGroup到Tracker的定义中。Event方法内，t.wg.Add(1) 被调用，这会增加计数器。一旦创建Goroutine，该Event函数就会返回，这满足了客户最小化时间跟踪延迟的要求。创建Goroutine开始工作，完成后会t.wg.Done()。Done减少了计数器，一边WaitGroup知道该Goroutine完成了

Add和Done对跟踪Goroutines活动数量很有用，但仍然必须只是程序等待它们完成。为了实现这一点，该Tracker类型声明一个新方法ShutDown，该函数最简单的实现是call t.wg.Wait(), 他会阻塞，知道goroutines计数减少到0，最后必须从主Goroutine中调用该方法

- **清单6**
```go

func main(){
	var a app

	a.track.ShutDown()
}
```
清单6主要是通过a.track.ShutDown()，阻塞main函数终止，直到a.track.ShutDown()完成

#### 但也许不要等太久

该ShutDown方法实现很简单并且可以完成所需工作，它等待goroutines完成。不幸的是，等待时间没有限制。根据您的生产环境，您可能不愿意无限期地等待程序关闭。为了给该方法添加截止日期，将其更改为：

- **清单7**
```go
func (t *Tracker)ShutDown(ctx context.Context) error{
	ch := make(chan struct{})

	go func(){
		t.wg.Wait()
		close(ch)
	}

	select {
	case <- ch:
		return nil
	case <- ctx.Done():
		return errors.New("timeout")
	}

}

```
在清单7中，该ShutDown方法采用一个context.Context作为输入，这就是Shutdown调用者允许等待时间的方式。在函数中创建一个ch通道，然后启动一个Goroutine，新Goroutine的唯一工作是等待WaitGroup完成，然后关闭通道，最后，使用一个select模块，等待上下文被取消或通道被关闭。

接下来，团队将调用更改func main

- **清单8**
```go
const timeout = 5 * time.Second
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

err := a.track.ShutDown(ctx)

```
清单8中，在main函数中创建了一个具有5s中上下文的参数，将此值创递给设置愿意等待的a.track.Shutdown事件限制


### **结论**
随着goroutines的引入，该服务器处理程序能够最大限度的减少跟踪事件的API客户端的延迟成本。go只需要使用关键字在后台运行这项工作就很容易，但该方案存在完整性问题。正确的做到这一点需要努力保证所有相关的goroutine在让程序关闭之前都已终止。  

**并发是一个有用的工具，但必须谨慎使用。**
