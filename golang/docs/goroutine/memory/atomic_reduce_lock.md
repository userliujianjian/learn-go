### 如何使用Atomic减少锁争用

### 本文基于Go1.14

Go提供了内存同步机制，例如通道或互斥体，有助于解决不同的问题。在共享内存的情况下，互斥体可以保护内存免受数据竞争的影响。然而，尽管存在两个互斥体，Go还通过包提供原子内存原语atomic以提高性能。在深入研究解决方案之前，让我们首先会到数据竞赛。  

#### **数据竞赛**  
当两个或多个goroutine同时访问同一内存位置并且其中有至少一个正在写入时，可能会发生数据竞争。虽然map具有防止数据竞争的本机机制，但简单结构没有任何机制，因此很容易收到数据竞争的影响。  

为了说明数据竞争，我将举一个由goroutine不断更新的配置示例。这是代码：  
```go
package main

import (
	"fmt"
	"sync"
)

type Config struct {
	a []int
}

func UpdateConfig() {
	cfg := &Config{}

	// Write
	go func() {
		i := 0
		for {
			i++
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				fmt.Println(cfg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func main() {
	UpdateConfig()
}

```  
运行此代码清楚表明，由于数据竞争，结果是不确定的.  
当结果相当随机时，每一行都期望是一个连续的整数序列。使用该标志运行相同的程序-race会指出数据竞争。  

```bash
&{[794238 796603 799254 801978 804696 807423]}
&{[810506 812857 815571 818288 821008 823724]}
&{[826573 829157 831889 834596 834598 837317]}
&{[840464 842755 845482 848195 850913 853635]}
&{[855185 856354 857700 857701 859065 861800]}
&{[864747 867272 870017 872666 875382 878125]}
&{[880957 883555 886264 888990 891693 894419]}
Found 7 data race(s)
exit status 66

```  
保护我们的读写免受数据竞争的影响可以通过互斥锁（可能是最常见）或包来完成atomic。 

#### **互斥与原子**  
标准库提供了两种互斥体的sync包： sync.Mutex和sync.RWMutex;当您的程序处理多个读取器和很少的写入器时，后者会得到优化。这是一种解决方案：  

```go
func updateConfigMutex() {
	cfg := &Config{}

	lock := sync.RWMutex{}

	// write
	go func() {
		var i int
		for {
			i++
			lock.Lock()
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
			lock.Unlock()
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				lock.RLock()
				fmt.Println(cfg)
				lock.RUnlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
```  

程序现在打印出预期结果；数字适当增加

```bash
&{[111 112 113 114 115 116]}
&{[112 113 114 115 116 117]}
&{[113 114 115 116 117 118]}

```  
借助atomic软件包可以完成第二种方案。这是代码：


```go
func updateByAtomic() {
	var v atomic.Value

	//writer
	go func() {
		var i int
		for {
			i++
			cfg := &Config{
				a: []int{i, i + 1, i + 2, i + 3, i + 4, i + 5},
			}
			v.Store(cfg)
		}
	}()

	// reader
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				cfg := v.Load()
				fmt.Println(cfg)
			}
			wg.Done()
		}()
	}

	wg.Wait()

}
```  
结果也在预料之中： 

```bash
&{[19948 19949 19950 19951 19952 19953]}
&{[19974 19975 19976 19977 19978 19979]}
&{[19978 19979 19980 19981 19982 19983]}
&{[20012 20013 20014 20015 20016 20017]}

```  

关于生成的输出，看起来使用该atomic包的解决方案要快得多，因为它可以生成更高的数字序列。对这两个程序进行基准测试有助于找出那一个更有效。  

#### 表现  
基准测试根据测量的内容来解释。在这种情况下，我将测量以前的程序，其中有一个不断存储新配置的编写器以及不断读取它的多个读取器。为了涵盖更多潜在的情况，假设配置不经常更改，我还将包括仅具有读取器的程序的基准测试。以下是新案例：   

```go
package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkMutexMultipleReaders(b *testing.B) {
	var lastValue uint64
	var lock sync.RWMutex

	cfg := Config{
		a: []int{0, 0, 0, 0, 0, 0},
	}

	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			for k := 0; k < 100; k++ {
				lock.RLock()
				atomic.SwapUint64(&lastValue, uint64(cfg.a[0]))
				lock.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
```


该基准测试证实了我们之前在性能方面所看到的情况。为了了解互斥体的平静到底在哪里，我们可以在启用追踪器的情况下冲洗运行程序。  
有关该trace包的更多信息，我建议您阅读[文章](https://medium.com/a-journey-with-go/go-discovery-of-the-trace-package-e5a821743c3c)  

这是使用该包的程序的配置文件atomic：  
[reduce_data_race-1.png](../img/reduce-data-race-1.png)  

Goroutines不间断地运行并且能够完成任务。关于带有互斥体的程序的配置文件，这是完全不同的：  

[reduce_data_race-2.png](../img/reduce-data-race-2.png)  

现在运行时间相当碎片化，这是由于停放goroutine的互斥体造成的。这可以从goroutine的该树种得到证实，其中显示了同步阻塞和异步阻塞话费的时间：  
[reduce_data_race-3.png](../img/reduce-data-race-3.png)  

阻塞时间大约占三分之一时间。从阻塞配置文件中可以详细了解：  
[reduce_data_race-4.png](../img/reduce-data-race-4.png)  

在这种情况下，该atomic软件包肯定会带来优势。然而，某些系统的性能可能会下降。例如： 如果您必须存储一张大地图，则每次更新地图时都必须肤质它，从而导致效率低下。  

有关互斥锁的更多信息，我建议您阅读这篇[文章](https://medium.com/a-journey-with-go/go-mutex-and-starvation-3f4f4e75ad50)

```go

```