### map如何实现get操作

下面是一个从map中获取key的示例：
```go 
func main() {
	ageMap := make(map[string]int)
	ageMap["hll"] = 18
	// 不带comma用法
	age1 := ageMap["lhh"]
	fmt.Println(age1) // 0

	// 带comma 用法
	age2, ok := ageMap["lhh"]
	fmt.Println(age2, ok) // 0 false
}
```
> Go读取map的方法有两种： 带 `comma` 、不带`comma`。

很好奇同一个函数有两种返回值是怎么实现的，这其实是编译器在背后做的工作：
  - 分析代码后，将两种语法对应底层两个不同的函数。
  ```go 
// src/runtime/hashmap.go
func mapaccess1(t *maptype, h *hmap, key unsafe.Pointer) unsafe.Pointer
func mapaccess2(t *maptype, h *hmap, key unsafe.Pointer) (unsafe.Pointer, bool)
  ```

另外根据key类型不同，编译器还会将查找、插入、删除的函数用更具体的函数替换掉，以优化效率
源文件：`src/runtime/hashmap_fast.go`

uint32 mapaccess1_fast32(....) unsafe.Pointer
uint32 mapaccess2_fast32(....) (unsafe.Pointer, bool)

参考文章：  
https://golang.design/go-questions/map/get/