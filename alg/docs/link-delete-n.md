## 删除倒数第N个节点

> 给你一个链表，删除链表的倒数第 n 个结点，并且返回链表的头结点。

**示例1**
![delete-node](../img/link_remove_n.png)
> 输入：head = [1,2,3,4,5], n = 2
> 输出：[1,2,3,5]

**示例2**
> 输入：head = [1], n = 1
> 输出：[]

**示例3**
> 输入：head = [1,2], n = 1
> 输出：[1]

**Tips**
- 链表中结点的数目为 sz
- 1 <= sz <= 30
- 0 <= Node.val <= 100
- 1 <= n <= sz

- 分析：
	- 审题：n为删除倒数第N个节点，这跟找到顺序链表中第N大的值是同一个问题
	- 方案：快慢指针、快慢指针加桶
	- 代码：code
	- 反馈：


- 答案（快慢指针）：  
- 解题思路：
	- 快慢指针法
		- 由于我们需要找到倒数第 nnn 个节点，因此我们可以使用两个指针first和second同时对链表进行遍历，**并且first比second超前 n 个节点**. 当first遍历到末尾时，second就恰好处于倒数第n个节点
		- 具体的：初始时first和second均指向头节点。我们线使用first进行遍历n次，**first和second之间间隔n-1个节点**，即first比second超前了n个节点（ps：一定要理解**间隔2个节点**相当于**超了3个节点**，也可以理解为3步）
		- 在这之后，我们同时使用first和second对链表进行便利。当first遍历到末尾（空指针），second恰好**指向**倒数第n个节点
		- 时间复杂度：O(L)，其中 L 是链表的长度。
		- 空间复杂度：O(1)。  
```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func removeNthFromEnd(head *ListNode, n int) *ListNode {
    if (head == nil) || (n == 1 && head.Next == nil) {
        return nil
    }
    // 初始first 和second均指向头节点，首先使用first对链表进行遍历，遍历次数为n，此时first和second之间间隔n-1个节点
    dummp := &ListNode{0, head}
    first, second := head, dummp
    // 先遍让快节点跑到n
    for i := 0; i < n; i++{
        first = first.Next
    }

    for ; first != nil; first = first.Next {
        second = second.Next
    }

    second.Next = second.Next.Next
    return dummp.Next
}
```

- 答案（栈）：
- 解题思路：
	- 我们也可以在便利链表的同时将所有节点依次入栈。根据先进后出原则，我们弹出n个节点就是需要删除的节点，并且目前栈顶的节点就是带删除节点的钱去节点。这样一来，删除操作就变的十分方便了（PS：**记得栈内每个元素都是相互连接的，删除前驱节点的后一个节点就可以了。考虑到删除第一个节点时没有前驱节点，那就给它加一个**）

- 复杂度：
	- 时间复杂度O(L), 其中L时链表长度
	- 空间复杂度：O(L), 其中L是链表长度，主要为栈开销

```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func removeNthFromEnd(head *ListNode, n int) *ListNode {
    if (head == nil) || (n == 1 && head.Next == nil) {
        return nil
    }
    // 初始first 和second均指向头节点，首先使用first对链表进行遍历，遍历次数为n，此时first和second之间间隔n-1个节点
    dummp := &ListNode{0, head}
    buckets := []*ListNode{}
    for node := dummp; node != nil; node = node.Next {
        buckets = append(buckets, node)
    }

    prev := buckets[len(buckets) - n - 1]
    prev.Next = prev.Next.Next
    return dummp.Next
}
```



