### 向右旋转链表

- 题目：给你一个链表的头节点 head ，旋转链表，将链表每个节点向右移动 k 个位置。
**示例1**
![examp1](../img/link_roate_right_1.png)
> 输入：head = [1,2,3,4,5], k = 2
> 输出：[4,5,1,2,3]

**示例2**
![examp1](../img/link_roate_right_1.png)
> 输入：head = [0,1,2], k = 4
> 输出：[2,0,1]

- 分析：链表向右移动k个位置相当于从倒数第k个位置反转

- 答案：
- 解题思路1（**FIFO队列**）：
	- 从题目中能看出，右移后从尾部节点去掉一个再放到头部，如果分开两步就是将多余的节点，放入一个FIFO队列中，然后依次取出指向头节点
```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

func getLength(list *ListNode) int {
    num := 0
    for list != nil {
        num += 1
        list = list.Next
    }  
    return num
}

func rotateRight(head *ListNode, k int) *ListNode {
    // 获取节点长度
    length := getLength(head)
    if length <= 1 || k %length < 1 {
        return head
    }
    dummp := head
    num := length - (k % length)
    buckets := []*ListNode{}
    for i :=1; i <= length; i++ {
        if i > num {
            buckets = append(buckets, head)
            head = head.Next
        }else if i == num {
            tmp := head.Next
            head.Next = nil
            head = tmp
        }else{
            head = head.Next
        }
    }

    for i := len(buckets) - 1; i >= 0; i -- {
        item := buckets[i]
        item.Next = dummp
        dummp = item 
    }
    return dummp
}
```

- 解题思路2（**闭合为环**）：
	- 整体思路就是先把链表形成一个环，找最后节点的时捎带计算出长度，然后找到新链表的最后一个元素（相当于链表长度-k%链表长度，避免循环）断开。 **主要有两点，1:实际要友移几次。2:寻找从第几个元素之后断开，从原来头部的前一个节点开始数**
```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */

func rotateRight(head *ListNode, k int) *ListNode {
    // 获取节点长度
    if head == nil || head.Next == nil {
        return head
    }

    length := 1
    cur := head
    for cur.Next != nil{
        cur = cur.Next
        length += 1
    }

    if length <= 1 || k %length < 1 {
        return head
    }
  
    num := length - (k % length)
    // 闭合环
    cur.Next = head
    // 移动num个位置后断开环
    for i := num; i > 0; i -- {
        cur = cur.Next
    }

    dummp := cur.Next
    cur.Next = nil
    
    return dummp
}
```
