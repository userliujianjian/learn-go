### 什么是csp
> Do not communicate by sharing memory; instead, share memory by communicating.  
翻译过来就是： 不要通过共享内存来通信，而要通过通信来共享内存。 

这就是Go的并发哲学，它依赖CSP模型，基于 Channel实现。  
CSP 经常被认为是 GO 在病发变成上成功的关键因素。 CSP全称是“Communicating Sequential Processes”, 这也是Tony Hoare 在1978年发表在ACM的一篇论文。论文里指出一门编程语言应该重视 input 和 output的原语，尤其是并发编程的代码。  
在那篇文章发表的时代，人们正在研究模块化编程思想，该不该用goto语句在当时是最激烈的议题。 彼时，面向对象的编程思想正在崛起，几乎没什么人关心并发编程。  
在文章中，CSP也是一门自定义的编程语言，作者定义了输入输出语句，用于processes间的通信（communication）。Processes被认为是需要输入驱动，并且产生输出，供其他processes消费，processes可以是进城、线程、甚至是代码块。输入命令是：!, 用来向process写入；输出时：？，用来从processes独处。这篇文章讲的channel正式借鉴了这一设计。

