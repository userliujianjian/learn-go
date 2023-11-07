### 切片的容量是怎么增长的

在项目开发过程中，使用切片类型变量时，有没有遇到过以下情况？
```go
func SliceCap() {
	fmt.Println("[SliceCap] is start...")
	num := make([]int, 3, 4)
	num1 := append(num, 1)
	num2 := append(num1, 2)
	num1[0] = 1
	num2[1] = 2
	fmt.Printf("[SliceCap] num: %v, cap: %d \n", num, cap(num))    // [SliceCap] num: [1 0 0], cap: 4
	fmt.Printf("[SliceCap] num1: %v, cap: %d \n", num1, cap(num1)) // [SliceCap] num1: [1 0 0 1], cap: 4
	fmt.Printf("[SliceCap] num2: %v, cap: %d \n", num2, cap(num2)) // [SliceCap] num2: [0 2 0 1 2], cap: 8

	fmt.Println("[SliceCap] is end...")
}

```

我们思考一下为什么`num2[1] = 2` 没有影响num1 和 num呢？（tips：观察切片初始容量大小）


说到slice扩容就离不开append函数，使用 append 可以向 slice 追加元素，实际上是往底层数组添加元素。  
但是底层数组的长度是固定的，如果索引 len-1 所指向的元素已经是底层数组的最后一个元素，就没法再添加了。  
这时，slice会迁移到新内存位置，新底层数组的长度也会增加，这样就可以放置新增的元素。
同时，为了未来可能发生在此`append`操作，新的底层数组的长度应该是多少呢，如果每次添加元素时都发生迁移，成本太高。
从上述例子中可以看出，新slice预留出多个buff，那么容量增长有什么规律呢？让我们做一个小测试：
#### go1.9
```go
func SliceCap() {
	fmt.Println("[SliceCap] is start...")
	s := make([]int, 0)
	oldCap := cap(s)

	for i := 0; i < 2048; i++ {
		s = append(s, i)

		newCap := cap(s)

		if newCap != oldCap {
			fmt.Printf("[SliceCap][%d -> %4d] cap = %-4d  |  after append %-4d  cap = %-4d\n", 0, i-1, oldCap, i, newCap)
			oldCap = newCap
		}
	}
	fmt.Println("[SliceCap] is end...")

	//[SliceCap] is start...
	//[SliceCap][0 ->   -1] cap = 0     |  after append 0     cap = 1
	//[SliceCap][0 ->    0] cap = 1     |  after append 1     cap = 2
	//[SliceCap][0 ->    1] cap = 2     |  after append 2     cap = 4
	//[SliceCap][0 ->    3] cap = 4     |  after append 4     cap = 8
	//[SliceCap][0 ->    7] cap = 8     |  after append 8     cap = 16
	//[SliceCap][0 ->   15] cap = 16    |  after append 16    cap = 32
	//[SliceCap][0 ->   31] cap = 32    |  after append 32    cap = 64
	//[SliceCap][0 ->   63] cap = 64    |  after append 64    cap = 128
	//[SliceCap][0 ->  127] cap = 128   |  after append 128   cap = 256
	//[SliceCap][0 ->  255] cap = 256   |  after append 256   cap = 512
	//[SliceCap][0 ->  511] cap = 512   |  after append 512   cap = 848
	//[SliceCap][0 ->  847] cap = 848   |  after append 848   cap = 1280
	//[SliceCap][0 -> 1279] cap = 1280  |  after append 1280  cap = 1792
	//[SliceCap][0 -> 1791] cap = 1792  |  after append 1792  cap = 2560
	//[SliceCap] is end...
}

```
看到了嘛？golang1.8版本之后，在原来slice容量oldcap 小于256的时候，新切片的容量newscap的确是oldcap的两倍。  
但是，当oldcap容量大于等于 256 的时候，情况就有变化了。当向 slice 中添加元素 512 的时候，老 slice 的容量为 512，之后变成了 848，两者并没有符合newcap = oldcap+(oldcap+3*256)/4 的策略（512+（512+3*256）/4）=832。添加完 848 后，新的容量 1280 当然也不是 按照之前策略所计算出的的1252。 让我们看源码：


##### golang版本 1.9
```go
func growslice(oldPtr unsafe.Pointer, newLen, oldCap, num int, et *_type) slice {
	oldLen := newLen - num
	// ...

	newcap := oldCap
	doublecap := newcap + newcap
	if newLen > doublecap {
		newcap = newLen
	} else {
		const threshold = 256
		if oldCap < threshold {
			newcap = doublecap
		} else {
			// Check 0 < newcap to detect overflow
			// and prevent an infinite loop.
			for 0 < newcap && newcap < newLen {
				// Transition from growing 2x for small slices
				// to growing 1.25x for large slices. This formula
				// gives a smooth-ish transition between the two.
				newcap += (newcap + 3*threshold) / 4
			}
			// Set newcap to the requested cap when
			// the newcap calculation overflowed.
			if newcap <= 0 {
				newcap = newLen
			}
		}
	}

	var overflow bool
	var lenmem, newlenmem, capmem uintptr
	// Specialize for common values of et.Size.
	// For 1 we don't need any division/multiplication.
	// For goarch.PtrSize, compiler will optimize division/multiplication into a shift by a constant.
	// For powers of 2, use a variable shift.
	switch {
	case et.Size_ == 1:
		lenmem = uintptr(oldLen)
		newlenmem = uintptr(newLen)
		capmem = roundupsize(uintptr(newcap))
		overflow = uintptr(newcap) > maxAlloc
		newcap = int(capmem)
	case et.Size_ == goarch.PtrSize:
		lenmem = uintptr(oldLen) * goarch.PtrSize
		newlenmem = uintptr(newLen) * goarch.PtrSize
		capmem = roundupsize(uintptr(newcap) * goarch.PtrSize)
		overflow = uintptr(newcap) > maxAlloc/goarch.PtrSize
		newcap = int(capmem / goarch.PtrSize)
	case isPowerOfTwo(et.Size_):
		var shift uintptr
		if goarch.PtrSize == 8 {
			// Mask shift for better code generation.
			shift = uintptr(sys.TrailingZeros64(uint64(et.Size_))) & 63
		} else {
			shift = uintptr(sys.TrailingZeros32(uint32(et.Size_))) & 31
		}
		lenmem = uintptr(oldLen) << shift
		newlenmem = uintptr(newLen) << shift
		capmem = roundupsize(uintptr(newcap) << shift)
		overflow = uintptr(newcap) > (maxAlloc >> shift)
		newcap = int(capmem >> shift)
		capmem = uintptr(newcap) << shift
	default:
		lenmem = uintptr(oldLen) * et.Size_
		newlenmem = uintptr(newLen) * et.Size_
		capmem, overflow = math.MulUintptr(et.Size_, uintptr(newcap))
		capmem = roundupsize(capmem)
		newcap = int(capmem / et.Size_)
		capmem = uintptr(newcap) * et.Size_
	}

	...
}

```

如果只看前半部分，现在网上各种文章里说的 newcap 的规律是对的。现实是，后半部分还对 newcap 作了一个内存对齐，这个和内存分配策略相关。进行内存对齐之后，新 slice 的容量是要 大于等于 按照前半部分生成的newcap。  
之后，向 Go 内存管理器申请内存，将老 slice 中的数据复制过去，并且将 append 的元素添加到新的底层数组中。  

最后向`growslice` 函数调用者返回一个新的slice，这个slice的长度并没有变化，而容量却增大了。

【引申1】
```go
func SliceCapPlug1() {
	// s只有一个元素[5]
	s := []int{5}
	// s扩容，容量变为2，[5,7]
	s = append(s, 7)

	// s扩容，容量变为4 【5，7，9】，注意，这时s长度为3，只有3个元素
	s = append(s, 9)

	// 由于s的底层数组仍然有空间，并不会扩容。
	// 这样底层数组变成[5,7,9,11], 注意此时 s = [5,7,9], 容量为4；
	// x = [5,7,9,11] 容量为4，这里s不变
	x := append(s, 11)

	// 这里还是s元素的尾部追加元素，由于s的长度为3，容量为4，
	// 所以直接在底层数组所因为3的地方填上12，结果 s = [5,7,9], y=[5,7,9,12], x=[5,7,9,12]
	// x和y的长度均为4，容量也均为4
	y := append(s, 12)
	fmt.Println(s, x, y)
}
```
> 这里要注意的是，append函数执行完后，返回的是一个全新的 slice，并且对传入的 slice 并不影响。



【引申2】
```go
func SliceCapPlug2() {
	s := []int{1, 2}
	s = append(s, 3, 4, 5)
	fmt.Printf("[SliceCapPlug2] len=%d, cap=%d \n", len(s), cap(s))
	//[SliceCapPlug2] len=5, cap=6 
}

```

如果按网上各种文章中总结的那样：小于原 slice 长度小于 256 的时候，容量每次增加 1 倍。添加元素 4 的时候，容量变为4；添加元素 5 的时候不变；添加元素 6 的时候容量增加 1 倍，变成 8。
这是错误的！我们来仔细看看，为什么会这样，上面growslice中switch case 调用了，roundupsize,源码再次搬出代码：

例子中 s 原来只有 2 个元素，len 和 cap 都为 2，append 了三个元素后，长度变为 5，容量最小要变成 5，即调用 growslice 函数时，传入的第三个参数应该为 5。即 cap=5。而一方面，doublecap 是原 slice容量的 2 倍，等于 4。满足第一个 if 条件，所以 newcap 变成了 5。

接着调用了 roundupsize 函数，传入 40。（代码中ptrSize是指一个指针的大小，在64位机上是8）

```go
// go 1.9.13 runtime/msize.go

// 我们再看内存对齐，搬出 roundupsize 函数的代码：
func roundupsize(size uintptr) uintptr {
	if size < _MaxSmallSize {
		if size <= smallSizeMax-8 {
			return uintptr(class_to_size[size_to_class8[divRoundUp(size, smallSizeDiv)]])
		} else {
			return uintptr(class_to_size[size_to_class128[divRoundUp(size-smallSizeMax, largeSizeDiv)]])
		}
	}
	if size+_PageSize < size {
		return size
	}
	return alignUp(size, _PageSize)
}


```

这个函数的参数依次是 元素的类型，老的 slice，新 slice 最小求的容量。


这是 Go 源码中有关内存分配的两个 slice。class_to_size通过 spanClass获取 span划分的 object大小。而 size_to_class8 表示通过 size 获取它的 spanClass。

#### 课后总结
- 切片扩容的策略是什么呢？ 

参考文章：
https://golang.design/go-questions/slice/grow/


