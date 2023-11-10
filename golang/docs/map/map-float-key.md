### float类型可以作为map的key嘛？
- 从语法上看是可以的。   
- 除slice、map、functions这几种类型外，其他类型都是可以的。具体包含：布尔值、数字、字符串、指针、通道借口类型、结构体、只包含上述类型的数组。这些类型的共同特征是支持：`==`、`！=`操作符。  
- 如果是结构体，只有hash后的值相等，才被认为是相同的key注意：很多字面值相等，hash出来的值不一定相等。
```go
func floatKey() {
	m := make(map[float64]int)

	m[1.4] = 1
	m[2.4] = 2
	m[math.NaN()] = 3
	m[math.NaN()] = 3

	for k, v := range m {
		fmt.Printf("[%v , %d] \n", k, v)

		// [1.4 , 1] 
		// [2.4 , 2] 
		// [NaN , 3] 
		// [NaN , 3] 
	}

	fmt.Printf("k: %v, v: %d \n", math.NaN(), m[math.NaN()])                                   // k: NaN, v: 0
	fmt.Printf("k: %v, v: %d \n", 2.400000000001, m[2.400000000001])                           // k: 2.400000000001, v: 0
	fmt.Printf("k: %v, v: %d \n", 2.4000000000000000000000001, m[2.4000000000000000000000001]) // k: 2.4, v: 2
}
```
例子中插入了4个Key：1.4，2.4, NaN, NaN 打印的时候也出现4个Key难道 Nan!=NaN?  
接着查询key的时候，发现NaN不存在，2.400000000001也不存在，而2.4000000000000000000000001存在，是不是有点诡异？

接着从汇编中发现如下事实：  
	当float64作为key的时候，要先将其转换成uint64,在插入key
	具体是通过 `Float64frombits` 函数完成的  
	```go 
	// Float64frombits returns the floating point number corresponding
	// the IEEE 754 binary representation b.
	func Float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }
	```
	也就是将浮点数表示称IEEE754规定的格式
我们再来看一个例子
```go		
package main

import (
	"fmt"
	"math"
)

func main() {
	m := make(map[float64]int)
	m[2.4] = 2

    fmt.Println(math.Float64bits(2.4)) // 4612586738352862003
	fmt.Println(math.Float64bits(2.400000000001)) // 4612586738352864255
	fmt.Println(math.Float64bits(2.4000000000000000000000001)) // 4612586738352862003
}
```
输出如下：
	4612586738352862003
	4612586738352864255
	4612586738352862003
转成十六进制：
	0x4003333333333333
	0x4003333333333BFF
	0x4003333333333333

现在清晰多了， 2.4和2.4000000000000000000000001经过 `math.Float64bits`函数转换后结果是一样的，自然而这在map看来，是同一个key

接下来再看看NaN(not a number)

```go
// NaN returns an IEEE 754 `not-a-number` value.
func Nan() float64 {return Float64frombits(uvnan)}

// uvnan定义为
uvnan = 0x7FF8000000000001

```
NaN() 直接调用Float64frombits, 传入写死的const变量 0x7FF8000000000001，得到NaN类型值。   
既然Nan是从一个常量解析得来的，那么为什么map时，会被认为不是同一个key呢？

这是由类型的哈希函数决定的，例如对于64为浮点数，他的哈希函数如下：

```go
func f64hash(p unsafe.Pointer, h uintptr) uintptr {
	f := *(*float64)(p)
	switch {
	case f == 0:
		return c1 * (c0 ^ h) // +0, -0
	case f != f:
		return c1 * (c0 ^ h ^ uintptr(fastrand())) // any kind of NaN
	default:
		return memhash(p, h, 8)
	}
}
```
**第二个case f!= f 就是针对NaN，这里会增加一个随机数 fastrand()**
**这样NaN存在两个key谜题就解开了。**

> **最后结论：float类型可以作为key，但是由于精度问题，会导致一些诡异的问题，慎用**










