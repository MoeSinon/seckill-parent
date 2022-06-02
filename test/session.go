package main

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

// func main() {

// 	slice := generateSlice(20)
// 	fmt.Println("\n--- Unsorted --- \n\n", slice)
// 	combsort(slice, 1.3, len(slice), len(slice))
// 	fmt.Println("\n--- Sorted ---\n\n", slice, "\n")
// }

// func main() {
// 	tree := &BinaryTree{}
// 	tree.insert(100).
// 		insert(-20).
// 		insert(-50).
// 		insert(-15).
// 		insert(-60).
// 		insert(50).
// 		insert(60).
// 		insert(55).
// 		insert(85).
// 		insert(15).
// 		insert(10).
// 		insert(5).
// 		insert(-10)
// 	print(os.Stdout, tree.root, 15, 'M')
// 	tree.del(15)
// 	print(os.Stdout, tree.root, 15, 'M')
// }

// func main() {
// 	link := List{}
// 	link.listincert(5)
// 	link.listincert(9)
// 	link.listincert(13)
// 	link.listincert(22)
// 	link.listincert(28)
// 	link.listincert(36)

// 	fmt.Println("\n==============================\n")
// 	fmt.Printf("Head: %v\n", link.head.key)
// 	fmt.Printf("Tail: %v\n", link.tail.key)
// 	fmt.Println("\n==============================\n")
// 	fmt.Printf("head: %v\n", link.head.key)
// 	fmt.Printf("tail: %v\n", link.tail.key)
// 	link.Reverse()
// 	fmt.Println("\n==============================\n")
// }

func main() {
	txt := "Australia is a country and continent surrounded by the Indian and Pacific oceans."
	patterns := []string{"and", "the", "surround", "Pacific", "Germany"}
	matches := Search(txt, patterns)
	fmt.Println(matches)
}

// Generates a slice of size, size filled with random numbers
func generateSlice(size int) []int {

	slice := make([]int, size, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		slice[i] = rand.Intn(999) - rand.Intn(999)
	}
	return slice
}

func combsort(items []int, shrink float64, gap, n int) {

	gap = int(float64(gap) / shrink)
	if gap == 0 {
		gap = 1
	}
	for i := 0; items[i] > items[i+gap] && i+gap < n; i++ {
		items[i+gap], items[i] = items[i], items[i+gap]
	}
	if gap == 1 {
		return
	}
	combsort(items, 1.3, gap, len(items))
}

type BinaryNode struct {
	left  *BinaryNode
	right *BinaryNode
	data  int64
}

type BinaryTree struct {
	root *BinaryNode
}

func (t *BinaryTree) insert(data int64) *BinaryTree {
	if t.root == nil {
		t.root = &BinaryNode{data: data, left: nil, right: nil}
	} else {
		t.root.insert(data)
	}
	return t
}

func (n *BinaryNode) insert(data int64) {
	if n == nil {
		return
	}
	if data <= n.data {
		if n.left == nil {
			n.left = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			n.left.insert(data)
		}
	} else {
		if n.right == nil {
			n.right = &BinaryNode{data: data, left: nil, right: nil}
		} else {
			n.right.insert(data)
		}
	}
}

func (n *BinaryTree) del(data int64) {
	if n == nil {
		return
	}
	tmp, tmp1 := n.search(data, n.root, n.root)
	if tmp1.left == tmp {
		tmp1.left = tmp.left
		if tmp.right != nil {
			n.insert(tmp.right.data)
		}
	} else {
		tmp1.right = tmp.right
		n.insert(tmp.left.data)
	}

}

func (n *BinaryTree) search(data int64, tmp, tmp1 *BinaryNode) (*BinaryNode, *BinaryNode) {
	if n == nil {
		return nil, nil
	}

	if data < tmp.data {
		tmp1 = tmp
		tmp = tmp.left
		return n.search(data, tmp, tmp1)

	}
	if data > tmp.data {
		tmp1 = tmp
		tmp = tmp.right
		return n.search(data, tmp, tmp1)

	}
	return tmp, tmp1
}

func print(w io.Writer, node *BinaryNode, ns int, ch rune) {
	if node == nil {
		return
	}
	for i := 0; i < ns; i++ {
		fmt.Fprint(w, "  ")
	}
	fmt.Fprintf(w, "%c:%v\n", ch, node.data)
	if node.left != nil {
		for i := 0; i < ns-2; i++ {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%c:%v", 'L', node.left.data)
	} else {
		fmt.Fprint(w, "\n")
	}

	if node.right != nil {
		for i := 0; i < ns-12; i++ {
			fmt.Fprint(w, "  .")
		}
		fmt.Fprintf(w, "%c:%v\n", 'R', node.right.data)
	} else {
		fmt.Fprint(w, "\n")
	}

	print(w, node.left, ns-2, 'L')
	print(w, node.right, ns+2, 'R')
}

type Node struct {
	prev *Node
	next *Node
	key  interface{}
}

type List struct {
	head *Node
	tail *Node
}

func (L *List) listincert(key interface{}) {
	list := &Node{
		next: L.head,
		key:  key,
	}
	if L.head != nil {
		L.head.prev = list
	}
	L.head = list

	l := L.head
	for l.next != nil {
		l = l.next
	}
	L.tail = l
}

func Display(list *Node) {
	for list != nil {
		fmt.Printf("%v ->", list.key)
		list = list.next
	}
	fmt.Println()
}

func ShowBackwards(list *Node) {
	for list != nil {
		fmt.Printf("%v <-", list.key)
		list = list.prev
	}
	fmt.Println()
}

func (l *List) Reverse() {
	curr := l.head
	var prev *Node
	l.tail = l.head

	for curr != nil {
		next := curr.next
		curr.next = prev
		prev = curr
		curr = next
	}
	l.head = prev
	Display(l.head)
}

const base = 16777619

func Search(txt string, patterns []string) []string {
	in := indices(txt, patterns)
	matches := make([]string, len(in))
	i := 0
	for j, p := range patterns {
		if _, ok := in[j]; ok {
			matches[i] = p
			i++
		}
	}

	return matches
}

func indices(txt string, patterns []string) map[int]int {
	n, m := len(txt), minLen(patterns)
	matches := make(map[int]int)

	if n < m || len(patterns) == 0 {
		return matches
	}

	var mult uint32 = 1 // mult = base^(m-1)
	for i := 0; i < m-1; i++ {
		mult = (mult * base)
	}

	hp := hashPatterns(patterns, m)
	h := hash(txt[:m])
	for i := 0; i < n-m+1 && len(hp) > 0; i++ {
		if i > 0 {
			h = h - mult*uint32(txt[i-1])
			h = h*base + uint32(txt[i+m-1])
		}

		if mps, ok := hp[h]; ok {
			for _, pi := range mps {
				pat := patterns[pi]
				e := i + len(pat)
				if _, ok := matches[pi]; !ok && e <= n && pat == txt[i:e] {
					matches[pi] = i
				}
			}
		}
	}
	return matches
}

func hash(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = (h*base + uint32(s[i]))
	}
	return h
}

func hashPatterns(patterns []string, l int) map[uint32][]int {
	m := make(map[uint32][]int)
	for i, t := range patterns {
		h := hash(t[:l])
		if _, ok := m[h]; ok {
			m[h] = append(m[h], i)
		} else {
			m[h] = []int{i}
		}
	}

	return m
}

func minLen(patterns []string) int {
	if len(patterns) == 0 {
		return 0
	}

	m := len(patterns[0])
	for i := range patterns {
		if m > len(patterns[i]) {
			m = len(patterns[i])
		}
	}

	return m
}
