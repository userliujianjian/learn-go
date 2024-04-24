### Golang的map实现原理

### 介绍
在写代码的过程中，经常会遇到函数之间传递参数的情况，我们都知道Golang函数参数是`值传递`。
但传递切片或map类型的参数，总会发现公共的变量会莫名其妙的被改动。难道是Golang的BUG嘛？这可能是一个非常小但容易致命的问题。

#### 未完成的工作
要查看未完成工作的简单示例，请检查此程序。

- 清单1
```go
package main

import "fmt"

// 函数的参数通过值传递-map
func initDataInfo(data map[string]string) {
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
    initDataInfo(info)
    fmt.Println("step2 info length", len(info))
}

```

- output:
```text
step1 info Length:  0
step2 info length 4
```

看完`清单1`中的代码，有个问题一直困惑着我。变量`info`传给初始化函数`initDataInfo`，根据`函数值传递规则`，`info`变量并不会被修改，那为什么
`step2 info length`的长度变成4了？ 带着这个疑问我们去看变量`info`类型底层的数据结构。

- 清单2
```go

```




