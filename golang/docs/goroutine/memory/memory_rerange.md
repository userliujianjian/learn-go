## 内存重排
### 简介
本文是抄写曹大谈内存重排的文章，抄写以便加深自己的理解，底部有原文连接。其中有一点讲，在阅读中英文参考资料时，发现英文的我能读懂，读中文却很费劲。经过对比，其实英文文章通常是由一个个例子引入，循序渐进，逐步深入。跟着坐着的脚步探索，非常有意思。而中文的博客上来就直奔主题，对新接触者很不友好。
- 什么是内存重排
	- CPU重排
	- 编译器重排
- 为什么要内存重排
- 内存重排的底层原理
- 总结
- 参考资料

#### 什么是内存重排  
氛围两种，硬件和软件层面的，包括CPU重排、编译器重排。