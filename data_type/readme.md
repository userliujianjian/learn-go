### 数据类型

- #### 切片
  切片（slice）是一个拥有相同元素类型的可变长度的序列。它是基于数组类型做的一层封装。非常灵活支持自动扩容。  

源码定义如下：
```go
type slice struct {
	array unsafe.Pointer // 元素指针
	len int // 实际元素个数
	cap int // 容量
}
```
#### 切片结构图。
![结构图](img/slice-data-type.png)

- ### 常见问题
#### Q1：切片与数组有什么区别?
> 数组是一个长度固定的数据类型，其长度在声明时就已经确定，不能动态改变；
> 切片是一个长度可变的数据类型，其长度在定义时可为空，也可以指定一个初始长度。

- 数组示例
```go
func Add(numbers [5]int){
    for i :=0; i < len(numbers); i++ {
        numbers[i] = numbers[i] + 1
    }
    // 如果数组作为函数的参数，那么实际传递的是一份数组的拷贝，而不是数组的指针。这也就意味着，在函数中修改数组的元素不会影响到原始数组
    fmt.Println("numbers in Add:", numbers) // [2,3,4,5,6]
}

func main(){
    var numbers [5]int
    for i := 0; i < len(numbers); i++ {
        numbers[i] = i + 1
    }
    Add(numbers)
    // 如果数组作为函数的参数，那么实际传递的是一份数组的拷贝，而不是数组的指针。这也就意味着，在函数中修改数组的元素不会影响到原始数组
    fmt.Println("numbers in main: ", numbers) // [1,2,3,4,5]
}
```
#### 切片
  切片（slice）是一个拥有相同元素类型的可变长度的序列。它是基于数组类型做的一层封装。非常灵活支持自动扩容。  
  源码定义如下：
  ```go
    type slice struct {
        array unsafe.Pointer // 指针
        len int // 实际元素个数
        cap int // 容量
    }
  ```

- 切片示例
```go
import "fmt"

func SliceVar() {
	var nums []int                    // 生命切片
	fmt.Println(len(nums), cap(nums)) // 0, 0
	nums = append(nums, 1)            // 初始化
	fmt.Println(len(nums), cap(nums)) // 1, 1

	nums1 := []int{1, 2, 3, 4}          // 生命并初始化
	fmt.Println(len(nums1), cap(nums1)) // 4 4

	nums2 := make([]int, 3, 5)          // 使用make函数构造切片
	fmt.Println(len(nums2), cap(nums2)) // 3 5

}

func SliceFunc(numbers []int) {
	for i := 0; i < len(numbers); i++ {
		numbers[i] = numbers[i] + 1
	}
	fmt.Println("numbers is SliceFunc: ", numbers) // [2 3 4 5 6]
}

func SliceMain() {
	var numbers []int
	for i := 0; i < 5; i++ {
		numbers = append(numbers, i+1)
	}
	SliceFunc(numbers)
  // 切片被当作参数传递时，直接修改的是切片本身
	fmt.Println("numbers in main: ", numbers) // [2 3 4 5 6]
}
```

### 总结
- 从长度讲：数组是一个长度固定的数据类型，定义时长度已经确定，不可动态改变；切片是一个长度可变的数据类型，定义时长度可为空，或用make函数指定初始长度
- 当作函数参数时，是否会改变原本的数据：函数操作的是数组的副本，不影响原数据；函数操作的时切片的引用，会影响原切片。
- 容量：切片有容量概念，指分配的内存空间

参考文章：
https://www.51cto.com/article/750465.html


