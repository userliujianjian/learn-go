### goroutine和线程的区别

#### 介绍
谈到goroutine，绕不开的一个话题：它和thread有什么区别？那我们从三个角度来讲：内存消耗，创建与销毁、切换。
> 分析：这是要考察GMP相关的知识，以及为什么goroutine可以无限开启。 此问题紧紧围绕着开销小，占用低来说

### **对比**
- 内存占用
	创建一个goroutine的内存消耗为2kB，实际运行过程中，如果栈空间不够用，会自动进行扩容。创建一个thread则需要消耗1MB栈内存，而且还需要备一个称为“a guard page”的区域用于和其他thread的栈空间进行隔离。  

- 创建和销毁
	Thread创建和销毁都会有巨大消耗，因为要和操作系统打交道，是内核级的，通常解决的方法就是线程池。而goroutine因为是由go runtime负责管理的，创建和销毁的消耗非常小，是用户级别

- 切换。
	当threads切换时，需要保存各种寄存器一边来恢复。
	而goroutine切换只需要三个寄存器：Program Counter, Stack Pointer and BP  
	一般而言，线程切换回消耗1000-1500纳秒，一个纳秒平均执行12-18条指令，所以由于线程切换，执行指令的条数会减少12000-18000。
	goroutine的切换约为200ns，想弹雨2400-3600条指令，因此goroutines切换成本要小的多。

> 当goroutine被调离cpu时，调度器负责把cpu寄存器的值保存在g对象的成员变量之中
> 当goroutine运行时， 调度器又负责把g对象的成员变量所保存的寄存器值恢复到cpu的寄存器


既然goroutine如此灵活，那么goroutine底层是如何实现的，让我们一起来看源码
```go
type g struct {

	// goroutine 使用的栈
	stack       stack   // offset known to runtime/cgo
	// 用于栈的扩张和收缩检查，抢占标志
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	_panic         *_panic // innermost panic - offset known to liblink
	_defer         *_defer // innermost defer
	// 当前与 g 绑定的 m
	m              *m      // current m; offset known to arm liblink
	// goroutine 的运行现场
	sched          gobuf
	syscallsp      uintptr        // if status==Gsyscall, syscallsp = sched.sp to use during gc
	syscallpc      uintptr        // if status==Gsyscall, syscallpc = sched.pc to use during gc
	stktopsp       uintptr        // expected sp at top of stack, to check in traceback
	// wakeup 时传入的参数
	param          unsafe.Pointer // passed parameter on wakeup
	atomicstatus   uint32
	stackLock      uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	goid           int64
	// g 被阻塞之后的近似时间
	waitsince      int64  // approx time when the g become blocked
	// g 被阻塞的原因
	waitreason     string // if status==Gwaiting
	// 指向全局队列里下一个 g
	schedlink      guintptr
	// 抢占调度标志。这个为 true 时，stackguard0 等于 stackpreempt
	preempt        bool     // preemption signal, duplicates stackguard0 = stackpreempt
	paniconfault   bool     // panic (instead of crash) on unexpected fault address
	preemptscan    bool     // preempted g does scan for gc
	gcscandone     bool     // g has scanned stack; protected by _Gscan bit in status
	gcscanvalid    bool     // false at start of gc cycle, true if G has not run since last scan; TODO: remove?
	throwsplit     bool     // must not split stack
	raceignore     int8     // ignore race detection events
	sysblocktraced bool     // StartTrace has emitted EvGoInSyscall about this goroutine
	// syscall 返回之后的 cputicks，用来做 tracing
	sysexitticks   int64    // cputicks when syscall has returned (for tracing)
	traceseq       uint64   // trace event sequencer
	tracelastp     puintptr // last P emitted an event for this goroutine
	// 如果调用了 LockOsThread，那么这个 g 会绑定到某个 m 上
	lockedm        *m
	sig            uint32
	writebuf       []byte
	sigcode0       uintptr
	sigcode1       uintptr
	sigpc          uintptr
	// 创建该 goroutine 的语句的指令地址
	gopc           uintptr // pc of go statement that created this goroutine
	// goroutine 函数的指令地址
	startpc        uintptr // pc of goroutine function
	racectx        uintptr
	waiting        *sudog         // sudog structures this g is waiting on (that have a valid elem ptr); in lock order
	cgoCtxt        []uintptr      // cgo traceback context
	labels         unsafe.Pointer // profiler labels
	// time.Sleep 缓存的定时器
	timer          *timer         // cached timer for time.Sleep

	gcAssistBytes int64
}
```

g结构体关联了两个比较简单的结构体，stack标识goroutine运行时的栈：
```go
type stack struct{
	// 栈顶，低地址
	lo uintptr
	// 栈低，搞地址
	hi uintptr
}
```

Goroutine运行时，光有栈还不行，至少还的包括PC、SP等寄存器，gobuf就保存了这些值：
```go
type gobuf struct{
	// 存储 rsp 寄存器的值
	sp uintptr
	// 存储rip寄存器的值
	pc uintptr
	//指向goroutine
	g guintptr
	ctxt unsafe.Pointer 
	// 保存系统调用的返回值
	ret sys.Uintreg
	lr uintptr
	bp uintptr // for GOEXPERIMENT=framepointer
}
```

