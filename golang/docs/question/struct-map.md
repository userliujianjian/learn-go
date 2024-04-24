### Golang的map实现原理

### 介绍
在写代码的过程中，经常会遇到函数之间传递参数的情况，我们都知道Golang函数参数是`值传递`。
但传递切片或map类型的参数，总会发现公共的变量会莫名其妙的被改动。难道是Golang的BUG嘛？这可能是一个非常小但容易致命的问题。

#### 未完成的工作
要查看未完成工作的简单示例，请检查此程序。

> 注：本文Golang版本为1.21.3  

- 清单1
```go
package main

import "fmt"

// 函数的参数通过值传递-map
func initMapData(data map[string]string) map[string]string {
    data["age"] = ""
    data["name"] = ""
    data["sex"] = ""
    data["school"] = ""
    return data
}

// 测试字典作为参数传递运行结果会如何
func main() {
    info := make(map[string]string)
    fmt.Println("step1 info Length: ", len(info))
    info = initMapData(info)
    fmt.Println("step2 info length", len(info))
}
```

- output:
```text
step1 info Length:  0
step2 info length 4
```

`清单1`中，`initMapData`函数负责初始化变量，并返回更新好的map字典。

- 清单2
```go
package main

import "fmt"


func initMapData(data map[string]string) map[string]string {
    data["age"] = ""
    data["name"] = ""
    data["sex"] = ""
    data["school"] = ""
    return data
}

// 测试字典作为参数传递运行结果会如何
func main() {
    info := make(map[string]string)
    fmt.Println("step1 info Length: ", len(info))
    info = initMapData(info)
    fmt.Println("step2 info length", len(info))

    infoV2 := make(map[string]string)
    fmt.Println("step3 infoV2 Length: ", len(infoV2))
    _ = initMapData(infoV2)
    fmt.Println("step4 infoV2 length", len(infoV2))
}
```   
- 输出：
```text 
step1 info Length:  0
step2 info length 4
step3 infoV2 Length:  0
step4 infoV2 length 4

```     
`清单2`跟`清单1`相比多了两步，变量`infoV2`调用`initMapData`进行初始化。不知道你有没有发现，这次调用`initMapData`函数，并没接收返回值，但变量`infoV2`的长度还是增加了等同于变量被修改。 细思一下`Golang函数参数为值传递`，但为什么`infoV2`却被改变了？  

让我们带着这个问题去看一下map的底层结构。  
- 清单3  
```go
type hmap struct {
    // Note: the format of the hmap is also encoded in cmd/compile/internal/reflectdata/reflect.go.
    // Make sure this stays in sync with the compiler's definition.
    count     int // # live cells == size of map.  Must be first (used by len() builtin)
    flags     uint8
    B         uint8  // log_2 of # of buckets (can hold up to loadFactor * 2^B items)
    noverflow uint16 // approximate number of overflow buckets; see incrnoverflow for details
    hash0     uint32 // hash seed

    buckets    unsafe.Pointer // array of 2^B Buckets. may be nil if count==0.
    oldbuckets unsafe.Pointer // previous bucket array of half the size, non-nil only when growing
    nevacuate  uintptr        // progress counter for evacuation (buckets less than this have been evacuated)

    extra *mapextra // optional fields
}
```  

`清单3`中结构体`hmap`就是map的底层数据结构，其中`buckets`,`oldbuckets`是指针格式，并且装map实际的数据。   
根据`函数参数是值传递`规则来看，当map作为参数时，给`initMapData`函数传递的确实是hmap的副本，`buckets`,`oldbuckets`等指针指值并未被改变。 所以在函数`initMapData`函数中，**新增的key实际上改变了`buckets`,`oldbuckets`指针指向存储的数据，地址并没哟被改变**。同理也就能理解为什么**清单2中infoV2即使没接收函数返回值，长度也变为4的原因**。  
最终我们的代码可以写成如下：
```go
package main

import "fmt"

func initMapData(data map[string]string) {
    data["age"] = ""
    data["name"] = ""
    data["sex"] = ""
    data["school"] = ""
    return
}

// 测试字典作为参数传递运行结果会如何
func main() {
    info := make(map[string]string)
    fmt.Println("step1 info Length: ", len(info))
    initMapData(info)
    fmt.Println("step2 info length", len(info))

    infoV2 := make(map[string]string)
    fmt.Println("step3 infoV2 Length: ", len(infoV2))
    initMapData(infoV2)
    fmt.Println("step4 infoV2 length", len(infoV2))
}

```
- 总结





