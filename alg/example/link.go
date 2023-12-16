package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

// 反转链表
func reverseListIterative(head *ListNode) (prev *ListNode) {
	curr := head

	for curr != nil {
		//保存下个指针
		nextTemp := curr.Next
		// 将下个指针指向上一个指针
		curr.Next = prev
		// 上个指针指向当前指针
		prev = curr
		// 当前之前移入下个指针
		curr = nextTemp

		//nextTemp := curr.Next
		//curr.Next, prev = prev, curr
		//// 当前指针后移
		//curr = nextTemp

	}
	return prev
}

// 合并两个有序链表
func mergeListNode(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	current := dummy
	for l1 != nil && l2 != nil {
		if l1.Val < l2.Val {
			current.Next = l1
			l1 = l1.Next
		} else {
			current.Next = l2
			l2 = l2.Next
		}
		current = current.Next
	}

	if l1 != nil {
		current.Next = l1
	} else {
		current.Next = l2
	}

	return dummy.Next
}

func printList(head *ListNode) {
	curr := head
	for curr != nil {
		fmt.Printf("%d -> ", curr.Val)
		curr = curr.Next
	}
	fmt.Println("nil")
}

func LinkInit() {
	list := &ListNode{1, &ListNode{2, &ListNode{3, nil}}}
	res := reverseListIterative(list)
	fmt.Println("链表反转 reverseListIterative 开始：")
	printList(res)
	fmt.Println("链表反转 reverseListIterative 结束")

	list1 := &ListNode{1, &ListNode{3, &ListNode{5, nil}}}
	list2 := &ListNode{2, &ListNode{4, &ListNode{6, nil}}}
	mergeList := mergeListNode(list1, list2)
	fmt.Println("合并两个有序链表 mergeListNode 开始：")
	printList(mergeList)

}
