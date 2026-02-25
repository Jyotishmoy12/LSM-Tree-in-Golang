package sstable

import (
	"os"
	"testing"
)

func TestSSTable_ReadWrite(t *testing.T) {
	path := "test_readwrite.sst"
	defer os.Remove(path)

	// 1. Write data
	w, _ := NewWriter(path)
	w.WritePair([]byte("apple"), []byte("red"), 0)
	w.WritePair([]byte("banana"), []byte("yellow"), 0)
	w.WritePair([]byte("grape"), []byte("purple"), 0)
	w.Close()

	// 2. Read data
	r, err := Open(path)
	if err != nil {
		t.Fatalf("Failed to open reader: %v", err)
	}
	defer r.Close()

	val, found, err := r.Get([]byte("banana"))
	if err != nil {
		t.Fatalf("Error during Get: %v", err)
	}
	if !found || string(val) != "yellow" {
		t.Errorf("Expected yellow, got %s", string(val))
	}

	// 3. Test non-existent key
	_, found, _ = r.Get([]byte("orange"))
	if found {
		t.Error("Should not have found 'orange'")
	}
}
