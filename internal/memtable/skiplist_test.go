package memtable

import (
	"bytes"
	"testing"
)

func TestSkipList_Basic(t *testing.T) {
	sl := NewSkipList()

	// Test Put
	sl.Put([]byte("apple"), []byte("red"))
	sl.Put([]byte("banana"), []byte("yellow"))
	sl.Put([]byte("cherry"), []byte("dark red"))

	// Test Get
	val, found := sl.Get([]byte("banana"))
	if !found || string(val) != "yellow" {
		t.Errorf("Expected yellow, got %s", string(val))
	}

	// Test Update
	sl.Put([]byte("apple"), []byte("green"))
	val, _ = sl.Get([]byte("apple"))
	if string(val) != "green" {
		t.Errorf("Expected green after update, got %s", string(val))
	}

	// Test Missing Key
	_, found = sl.Get([]byte("dragonfruit"))
	if found {
		t.Error("Expected not found for missing key")
	}
}

func TestSkipList_Ordering(t *testing.T) {
	sl := NewSkipList()
	sl.Put([]byte("z"), []byte("1"))
	sl.Put([]byte("a"), []byte("2"))
	sl.Put([]byte("m"), []byte("3"))

	// Manual traversal of level 0 (the full linked list)
	curr := sl.head.next[0]
	var lastKey []byte
	for curr != nil {
		if lastKey != nil && bytes.Compare(lastKey, curr.key) >= 0 {
			t.Errorf("Keys out of order: %s then %s", lastKey, curr.key)
		}
		lastKey = curr.key
		curr = curr.next[0]
	}
}
