package memtable

import (
	"bytes"
	"math/rand"
)

const (
	MaxLevel = 12  // Maximum height of the skip list
	P        = 0.5 // Probability factor for increasing level
)

// Node represents a single element in the SkipList
type Node struct {
	key   []byte
	value []byte
	next  []*Node // Array of pointers to next nodes at different levels
}

// Iterator allows for sequential traversal of the SkipList.
type Iterator struct {
	curr *Node
}

// SkipList is the sorted in-memory structure
type SkipList struct {
	head  *Node
	level int // Current highest level
}

// Key is a getter for the unexported key field
func (n *Node) Key() []byte {
	return n.key
}

// Value is a getter for the unexported value field
func (n *Node) Value() []byte {
	return n.value
}

// Next returns the next node at level 0 (for iteration)
func (n *Node) Next() *Node {
	return n.next[0]
}

// NewSkipList initializes a new SkipList with a dummy head node
func NewSkipList() *SkipList {
	return &SkipList{
		head:  &Node{next: make([]*Node, MaxLevel)},
		level: 0,
	}
}

// randomLevel decides how many levels a new node will occupy
func randomLevel() int {
	lvl := 0
	for rand.Float64() < P && lvl < MaxLevel-1 {
		lvl++
	}
	return lvl
}

// Put inserts or updates a key-value pair
func (s *SkipList) Put(key, value []byte) {
	update := make([]*Node, MaxLevel)
	curr := s.head

	// 1. Find the position for the new node
	for i := s.level; i >= 0; i-- {
		for curr.next[i] != nil && bytes.Compare(curr.next[i].key, key) < 0 {
			curr = curr.next[i]
		}
		update[i] = curr
	}

	curr = curr.next[0]

	// 2. If key exists, update value
	if curr != nil && bytes.Equal(curr.key, key) {
		curr.value = value
		return
	}

	// 3. Otherwise, insert new node with random level
	lvl := randomLevel()
	if lvl > s.level {
		for i := s.level + 1; i <= lvl; i++ {
			update[i] = s.head
		}
		s.level = lvl
	}

	newNode := &Node{
		key:   key,
		value: value,
		next:  make([]*Node, lvl+1),
	}

	for i := 0; i <= lvl; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}
}

// Get retrieves a value by key
func (s *SkipList) Get(key []byte) ([]byte, bool) {
	curr := s.head
	for i := s.level; i >= 0; i-- {
		for curr.next[i] != nil && bytes.Compare(curr.next[i].key, key) < 0 {
			curr = curr.next[i]
		}
	}
	curr = curr.next[0]

	if curr != nil && bytes.Equal(curr.key, key) {
		return curr.value, true
	}
	return nil, false
}

// NewIterator returns an iterator for the SkipList
func (s *SkipList) NewIterator() *Iterator {
	return &Iterator{curr: s.head.next[0]}
}

// next moves the next node and return the key and value of the current node

func (it *Iterator) Next() bool {
	if it.curr != nil && it.curr.next[0] != nil {
		it.curr = it.curr.next[0]
		return true
	} else {
		return false
	}
}

// Key returns the key of the current node
func (it *Iterator) Key() []byte {
	return it.curr.key
}

// Value returns the value of the current node
func (it *Iterator) Value() []byte {
	return it.curr.value
}
