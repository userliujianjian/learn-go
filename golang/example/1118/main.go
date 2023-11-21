package main

func main() {

	// 二叉树的深度优先、广度优先
	//1
	//2		3
	//4 	5   6 	7

	obj := &Data{}
	var nums []int16
	for {
		// P岸段做节点
		if obj.LeftNode != nil {
			nums = append(nums, obj.Value)
			// 获取左节点所有值
			getValue(obj)
		}
	}

}

// 获取
func getValue(obj *Data) int16 {
	if obj.LeftNode == nil {
		return obj.Value
	}
	return getValue(obj)
}

type Data struct {
	Value     int16
	LeftNode  *Data
	RightNode *Data
}
