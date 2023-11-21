package ch

//type hchan struct {
//	// chan 元素数量
//	qcount   uint
//	dataqsiz uint
//	// 指向底层循环数组的指针
//	//只针对有缓冲的 channel
//	buf unsafe.Pointer
//	//chan 中元素大小
//	elemsize uint16
//	// chan 是否被关闭的标志
//	closed uint32
//	// chan中的元素类型
//	elemtype *_type // element type
//	// 已发送元素在循环数组中的索引
//	sendx uint // send index
//	// 已接收元素在循环数组中的索引
//	recvx uint // receive index
//	// 等待接收的goroutine 队列
//	recvq waitq // list of recv waiters
//	// 等待发送的goroutine队列
//	sendq waitq // list of send waiters
//
//	// 保护hchan中所有字段
//	lock mutex
//}
