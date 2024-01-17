### 数据类型

### 切片与数组
- [数组与切片有什么异同](docs/slice/slice-array.md)
- [切片的容量是怎么增长的](docs/slice/slice-cap.md)
- [切片作为函数参数](docs/slice/slice-param.md)

### 哈希map
- [map实现原理](docs/map/map-create.md)
- [如何实现get操作](docs/map/map-get.md)
- [float类型可以作为map的key嘛](docs/map/map-float-key.md)
- [常见问题](docs/map/map.md)


### 接口
- [GO语言与鸭子类型的关系](docs/inter/duck.md)
- [值接受这和指针接受着的区别](docs/inter/pointer.md)
- [iface和eface的区别是什么](docs/inter/face.md)


### 通道
- [什么是CSP](docs/ch/csp.md)
- [channel的底层数据结构是什么](docs/ch/channel.md)
- [Goroutine泄漏](docs/ch/channel-buf.md)
- [Goroutine泄漏2](docs/goroutine/concurrency-trap-2.md)
- [Goroutine和GOMAXPROCS](docs/goroutine/goroutines-and-gomaxprocs.md)
- [并发模型之goroutine的使用时机](docs/goroutine/concurrency.md)
- [并发模型之谁发起谁负责](docs/goroutine/concurrency-caller.md)
- [并发模型之开启和关闭](docs/goroutine/concurrency-stop.md)
- [通道行为哲学](docs/ch/channel_behavior.md)
- [Go有缓冲和无缓冲通道](docs/ch/ch_buffer_unbuffer.md)
- [通道的本质](docs/ch/ch_nature.md)

### 语法
- [在Select语句中排序](docs/ch/go_select_order.md)
- [select常见错误](docs/ch/channel_select_bug.md)


### 内存（待进一步理解）
- [曹大谈内存重排](docs/goroutine/memory/memory_rerange.md)
- [内存重排](docs/goroutine/memory/memory_reordering.md)
- [通过通信共享内存](docs/goroutine/memory/memory_communicating.md)
- [如果对齐的内存写入是原子，为什么使用sync/atomic](docs/goroutine/memory/memory_aligned.md)
- [冰激淋制造商和数据竞争](docs/goroutine/memory/data_race_ice_cream.md)
- [冰激淋制造商和数据竞争2](docs/goroutine/memory/data_race_ice_cream-2.md)
- [如何处理数据竞争](docs/goroutine/memory/atomic_reduce_lock.md)
- [发现跟踪包](docs/goroutine/memory/discovery_trace_package.md)
- [互斥锁和饥饿模式](docs/goroutine/memory/mutex_and_starvation.md)

- [Go竞争检测器（理解68%）](docs/goroutine/memory/go_race_detector.md)
- [Go内存模型（未理解）](docs/goroutine/memory/mem.md)
- [Go内存屏障（未理解）](docs/goroutine/memory/memory_barrier.md)

### 场景
- [Go高级并发](docs/scenes/concurrency_advanced.md)


### 标准库
- [context使用](docs/lib/context.md)
- [context源码](docs/lib/context-2.md)

### 解释器
- [GMP](docs/schedule/gmp.md)

### 垃圾回收
- [GC垃圾回基础概念](docs/gc/gc.md)

### 常见面试题
- [概念题](docs/ch/ch-question.md)