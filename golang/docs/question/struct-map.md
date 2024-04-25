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

- 清单4  
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
一维字典传给函数`initMapData` 初始化，在函数内对字典进行更新，会同时更新函数外的变量值。
那二维的字典，在函数内更新，会不会影响到函数外的变量呢？

- 清单5  
```go   
package main

import (
    "fmt"
)

func initMapData(data map[string]string) {
    data["age"] = ""
    data["name"] = ""
    data["sex"] = ""
    data["school"] = ""
    return
}

func initTwoDimensional(data map[string]map[string]string) {
    data["v1"] = make(map[string]string)
    data["v1"]["age"] = ""
    data["v1"]["name"] = ""
    data["v1"]["sex"] = ""
    data["v1"]["school"] = ""
    data["v2"] = make(map[string]string)
    data["v2"]["age"] = ""
    data["v2"]["name"] = ""
    data["v2"]["sex"] = ""
    data["v2"]["school"] = ""
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

    //dimensional
    twoDimensional := make(map[string]map[string]string)
    fmt.Println("step5 twoDimensional Length: ", len(twoDimensional))
    initTwoDimensional(twoDimensional)
    fmt.Println("step6 twoDimensional Length: ", len(twoDimensional))
    for k, v := range twoDimensional {
        fmt.Println("step7 twoDimensional key: ", k, "length: ", len(v))
    }

}
```   
输出： 
```text
step1 info Length:  0
step2 info length 4
step3 infoV2 Length:  0
step4 infoV2 length 4
step5 twoDimensional Length:  0
step6 twoDimensional Length:  2
step7 twoDimensional key:  v1 length:  4
step7 twoDimensional key:  v2 length:  4

```   
从`清单5`输出的step6、step7中可以看出，一维的key `v1、v2`下的键值都被同步到源变量`twoDimensional`上了。 

- 总结  
本文主要围绕`**Golang函数参数是值传递**`规则，将字典作为参数，做了几个实验。__虽然在函数中修改字典变量，会影响到函数外的变量__ ，**但依然符合Golang是值传递的规则** , 之所以被改变是因为字典本身数据不在 `map`的结构上，`hmap`只是存储的真实数据对应的地址(`buckets`,`oldbuckets`). 感兴趣的话，大家可以继续看看Golang map的底层数据结构。 下一篇我依然会围绕`值传递`的话题，去看看切片的表现。  




