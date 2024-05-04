## Go值传递之切片(slice)
### 介绍  
上一篇文章介绍了Golang`值传递`中字典的表现，这一篇我们接着围绕`值传递`话题看看切片在实际代码中的表现。  

> 本文的Golang版本问1.21.3  

- 清单1:  
```go
func updateSliceItem(data []string, index int, val string) {
	data[index] = val
}

func main() {
	data := []string{"LJJ", "Bob", "Jon"}
	updateSliceItem(data, 0, "zhangsan")
	fmt.Println(data)
}
```  
- output输出：  
```text
step 1 data:  [LJJ Bob Jon]
step 2 修改后的数据 data: [zhangsan Bob Jon]
```  

`清单1`中将切片传递给函数`updateSliceItem`，修改指定下标元素的值，在未接收函数返回值的情况下，   
第0个元素被改变，这貌似不符合`值传递`的规则，这是个值得思考的问题。 我们来看一下切片的底层数据结构  

- slice切片数据结构：
```go 
// runtime/slice.go
type slice struct {
	array unsafe.Pointer // 元素指针
	len   int // 长度
	cap   int // 容量
}
```  
- 总结
切片底层数据存在`array`对应的指针上。那清单1中第0个元素在函数内部的修改影响到变量本身也就很好解释。 函数参数的data实际是slice的副本，但`array`的指针都指向同一个地址，所以在`函数内改变元素时会同时改变外部变量值`。  
**所以切片在函数中也是值传递**  

- 扩展阅读：
	- [切片](../question/struct-slice.md)
	- [切片作为参数](../slice/slice-param.md)
