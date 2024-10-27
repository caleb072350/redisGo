package list

import "redisGo/utils"

type LinkedList struct {
	first *node
	last  *node
	size  int
}

type node struct {
	val  interface{}
	prev *node
	next *node
}

func (list *LinkedList) Add(val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	n := &node{val: val}
	if list.last == nil {
		list.first = n
		list.last = n
	} else {
		list.last.next = n
		n.prev = list.last
		list.last = n
	}
	list.size++
}

func (list *LinkedList) find(index int) (val *node) {
	if index < list.size/2 {
		n := list.first
		for i := 0; i < index; i++ {
			n = n.next
		}
		return n
	} else {
		n := list.last
		for i := list.size - 1; i > index; i-- {
			n = n.prev
		}
		return n
	}
}

func (list *LinkedList) Get(index int) (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bounds")
	}
	return list.find(index).val
}

func (list *LinkedList) Set(index int, val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bounds")
	}
	list.find(index).val = val
}

func (list *LinkedList) Insert(index int, val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index > list.size {
		panic("index out of bounds")
	}
	if index == list.size {
		list.Add(val)
		return
	} else {
		// 找到index位置的元素，把新元素插在index位置上，原来的元素往后移
		pivot := list.find(index)
		n := &node{val: val, prev: pivot.prev, next: pivot}
		if pivot.prev == nil {
			list.first = n
		} else {
			pivot.prev.next = n
		}
		pivot.prev = n
		list.size++
	}
}

func (list *LinkedList) removeNode(n *node) {
	if n.prev == nil {
		list.first = n.next
	} else {
		n.prev.next = n.next
	}
	if n.next == nil {
		list.last = n.prev
	} else {
		n.next.prev = n.prev
	}

	// for gc
	n.prev = nil
	n.next = nil
	// n.val = nil // 这里因为删除之后要返回节点的值，所以不能将val设为nil
	list.size--
}

func (list *LinkedList) Remove(index int) (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if index < 0 || index >= list.size {
		panic("index out of bounds")
	}
	n := list.find(index)
	list.removeNode(n)

	return n.val
}

func (list *LinkedList) RemoveLast() (val interface{}) {
	if list == nil {
		panic("list is nil")
	}
	if list.last == nil {
		// empty list
		return nil
	}
	n := list.last
	list.removeNode(n)
	return n.val
}

func (list *LinkedList) RemoveAllByVal(val interface{}) int {
	if list == nil {
		panic("list is nil")
	}
	count := 0
	n := list.first
	for n != nil {
		var toRemoveNode *node
		if utils.Equals(n.val, val) {
			toRemoveNode = n
		}
		if n.next == nil {
			if toRemoveNode != nil {
				list.removeNode(toRemoveNode)
				count++
			}
			break
		} else {
			n = n.next
		}
		if toRemoveNode != nil {
			list.removeNode(toRemoveNode)
			count++
		}
	}
	return count
}

/**
 * remove at most `count` values of the specified value in this list scan from left to right
 */
func (list *LinkedList) RemoveByVal(val interface{}, count int) int {
	if list == nil {
		panic("list is nil")
	}
	if count <= 0 {
		return 0
	}
	n := list.first
	c := 0
	for i := 0; i < list.size; i++ {
		if utils.Equals(n.val, val) {
			list.removeNode(n)
			c++
			if c == count {
				break
			}
		}
		n = n.next
	}
	return c
}

func (list *LinkedList) ReverseRemoveByVal(val interface{}, count int) int {
	if list == nil {
		panic("list is nil")
	}
	if count <= 0 {
		return 0
	}
	n := list.last
	c := 0
	for i := list.size - 1; i >= 0; i-- {
		if utils.Equals(n.val, val) {
			list.removeNode(n)
			c++
			if c == count {
				break
			}
		}
		n = n.prev
	}
	return c
}

func (list *LinkedList) Len() int {
	if list == nil {
		panic("list is nil")
	}
	return list.size
}

func (list *LinkedList) ForEach(consumer func(int, interface{}) bool) {
	if list == nil {
		panic("list is nil")
	}
	n := list.first
	for i := 0; i < list.size; i++ {
		if !consumer(i, n.val) {
			break
		}
		n = n.next
	}
}

func (list *LinkedList) Contains(val interface{}) bool {
	contains := false
	list.ForEach(func(_ int, actual interface{}) bool {
		if actual == val {
			contains = true
			return false
		}
		return true
	})
	return contains
}

func (list *LinkedList) Range(start int, stop int) []interface{} {
	if list == nil {
		panic("list is nil")
	}
	if start < 0 || start >= list.size || stop < start || stop > list.size {
		panic("index out of bounds")
	}
	vals := make([]interface{}, stop-start)
	n := list.find(start)
	for i := start; i < stop; i++ {
		vals[i-start] = n.val
		n = n.next
	}
	return vals
}

func Make(vals ...interface{}) *LinkedList {
	list := &LinkedList{}
	for _, val := range vals {
		list.Add(val)
	}
	return list
}

func MakeBytesList(vals ...[]byte) *LinkedList {
	list := LinkedList{}
	for _, v := range vals {
		list.Add(v)
	}
	return &list
}
