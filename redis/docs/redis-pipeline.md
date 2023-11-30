### Redis pipeline

- 分析
	- Redis Pipeline 的面试主要停留在为什么 Pipeline 快这个核心要点，围绕着这个核心要点考察 Pipeline 的原理。
	- 同时还会结合考察如何在 Redis Cluster 里面使用 Pipeline，以及和批量命令的区别。


#### **Redis Pipeline 的原理**
- 分析：
	- 这里基本上就是回答 Pipeline 的实现机制
	- 在答基本步骤的过程中，可以有意识引导面试官问 N 的取值，N 对内存的影响
- 答案（Redis Pipeline 的原理是）：
	- 应用代码会持续不断的把请求发给 Redis Client；
	- Redis Client 会缓存这些命令，等凑够了 N 个，就发送命令到 Redis 服务端。而 N 的取值对 Pipeline 的性能影响比较大
	- Redis Server 收到命令之后进行处理，并且在处理完这 N 个命令之前，所有的响应都被缓存在内存里面。这里也可以看到，N 如果太大也会额外消耗 Redis Server 的内存
	- Redis Server 处理完了 Pipeline 发过来的一批命令，而后返回响应给 Redis Client；
	- Redis Clinet 接收响应，并且将结果递交给应用代码；
	- 如果此时还有命令，Redis Client 会继续发送剩余命令；

	亮点：Redis Pipeline 减少了网络 IO，也减少了 RTT，所以性能比较好。

- 类似问题
	- 为什么 Redis Pipeline 在实时性上要差一点？主要就是命令和响应都会被缓存，而不是及时返回。

#### **Redis Pipeline 有什么优势？**
- 分析：
	- 如果直接回答性能比较好，那么就基本等于没说。这个问题本质上其实是“为什么 Redis Pipeline 性能好”。
	- 结合之前我们的分析，可以看到无非就是两个原因：网络 IO 和 RTT。
- 答案：
	- Redis Pipeline 相比普通的单个命令模式，性能要好很多。
	- 单个命令执行的时候，需要两次 read 和 两次 send 系统调用，加上一个 RTT。如果有 N 个命令就是分别乘以 N。
	- 但是在 Pipeline 里面，一次发送，不管 N 多大，都是两次 read 和两次 send 系统调用，和一次 RTT。因而性能很好。


#### **Redis Pipeline 和 mget 的不同点**
- 分析：
	- 相同点：
		- 减少网络IO和RTT，性能好
		- Redis Cluster对这两种用法都不太友好
	- 不同点：
		- Redis Pipeline 可以执行任意的命令，而 mget 之类的只能是执行同种命令；
		- Redis Pipeline 的命令和响应都会被缓存，因此实时响应上不如 mget；
		- Redis Pipeline 和 mget 都会受到批次大小的影响，但是相比之下 Redis Pipeline 更加严重，因为它消耗内存更多；
- 答案：Redis Pipeline 和 mget 之类的批量命令有很多地方都很相似，比如说
	- 减少网络 IO 和 RTT，性能好（注意，这里可能面试官会问，它是如何减少 IO 和 RTT，也就是我们前面讨论优势的地方）
	- Redis Cluster 对这两种用法都不太友好（这个是引导，准备讨论 Redis Cluster 需要的特殊处理）

	- 命令种类：Redis Pipeline 可以执行任意的命令，而 mget 之类的只能是执行同种命令；
	- 是否缓存：Redis Pipeline 的命令和响应都会被缓存，因此实时响应上不如 mget；
	- N的大小是否影响性能： Redis Pipeline 和 mget 都会受到批次大小的影响，但是相比之下 Redis Pipeline 更加严重，因为它需要缓存命令和响应，消耗更大

	在频繁读写的情况下，使用 Redis Pipeline 都是能受益的。但是如果是追求实时响应的话，那么就不要使用 Redis Pipeline，因为 Redis Pipeline 的机制导致请求和响应会被缓存一小段时间。这种实时场景，只能考虑批量处理命令

- 类似问题
	- 什么时候选择 Redis Pipeline
	- 什么时候选择 mget
	
参考文章：https://github.com/flycash/interview-baguwen/blob/main/redis/pipeline.md
