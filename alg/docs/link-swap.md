### 两两交换

- 给你一个链表，两两交换其中相邻的节点，并返回交换后链表的头节点。你必须在不修改节点内部的值的情况下完成本题（即，只能进行节点交换）。

**示例1:**
![两两交换](../img/link_swap.png)

> 输入：head = [1,2,3,4]
> 输出：[2,1,4,3]

**示例2:**
> 输入：head = []
> 输出：[]

**示例3:**
> 输入：head = [1]
> 输出：[1]

- **tips**
	- 链表中节点的数目在范围 [0, 100] 内
	- 0 <= Node.val <= 100	

- 分析：
	- 根据切题四件套。
		- 审题：当元素个数小于2的时候原样返回，偶数个元素亮亮交换
		- 解决方案（两种）：递归、迭代
		- 写代码
		- 总结思考,feedback 
	- 解题思路：
		- 递归最重要的三点：返回值、调用单元做什么（重复做）、终止条件
			- 返回值：交换完的子链
			- 调用单元：设需要交换的两个点为head和next，head后面连接为交换完的子链，next连接head，完成交换
			- 终止条件：head或next为空指针，也就是当前节点为空或没有下一个节点，无法进行交换（非偶数个）
- 答案（递归）：
```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
 func swapPairs(head *ListNode) *ListNode {
    // 排除空和只有一个元素的情况
    if head == nil || head.Next == nil{
        return head
    }
    next := head.Next
    head.Next = swapPairs(next.Next)
    next.Next = head

    return next
}
```

-  答案（迭代）
	- 考虑小为什么需要加一个dummp和temp？
		- 当进行第二波调换3和4的时候，需要当前节点是2，所以dummp会在链之前加一个节点， temp是为了留一个变量指向头节点，返回时用
```go
/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func swapPairs(head *ListNode) *ListNode {
    // 排除空和只有一个元素的情况
    if head == nil || head.Next == nil{
        return head
    }
    dummp := &ListNode{Val: 0, Next:head}
    temp := dummp
    for temp.Next != nil && temp.Next.Next != nil{
        start := temp.Next
        end := temp.Next.Next
        // temp 的下一个指针之所以要指向第二个值是要跟当前链连接起来，否则会丢失
        temp.Next, start.Next, end.Next = end, end.Next, start
        temp = start
    }

    return dummp.Next
}
```