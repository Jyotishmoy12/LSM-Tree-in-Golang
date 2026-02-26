package memtable

import (
	"os"
	"testing"
)

func TestMemTable_PutAndGet(t *testing.T) {
	walPath := "test_memtable.wal"
	defer os.Remove(walPath)

	// Initialize with a 1KB limit
	mt, err := NewMemTable(walPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create MemTable: %v", err)
	}
	defer mt.Close()

	key := []byte("db_engine")
	val := []byte("lsm_tree")

	// Test Writing
	if err := mt.Put(key, val); err != nil {
		t.Errorf("Put failed: %v", err)
	}

	// Test Reading
	retrieved, found := mt.Get(key)
	if !found || string(retrieved) != "lsm_tree" {
		t.Errorf("Expected lsm_tree, got %s", string(retrieved))
	}

	// Test Fullness
	if mt.IsFull() {
		t.Error("MemTable should not be full yet")
	}
}
