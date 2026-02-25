package internal

import (
	"os"
	"testing"
)

func TestLSM_Basic(t *testing.T) {
	dir := "storage_test"
	defer os.RemoveAll(dir)

	lsm, err := New(dir, 1024) // 1KB threshold
	if err != nil {
		t.Fatalf("Failed to init LSM: %v", err)
	}
	defer lsm.Close()

	// Test Memory Read/Write
	lsm.Put([]byte("hero"), []byte("Batman"))
	val, found, _ := lsm.Get([]byte("hero"))
	if !found || string(val) != "Batman" {
		t.Errorf("Expected Batman, got %s", string(val))
	}
}

func TestLSM_FlushPersistence(t *testing.T) {
	dir := "storage_flush_test"
	defer os.RemoveAll(dir)

	// Set a very small max size to trigger a flush quickly (50 bytes)
	lsm, _ := New(dir, 50)
	
	// These writes should exceed 50 bytes and trigger flush()
	lsm.Put([]byte("key1"), []byte("value_that_is_quite_long_1"))
	lsm.Put([]byte("key2"), []byte("value_that_is_quite_long_2"))

	// Check if data is still there (it should now be on disk)
	val, found, _ := lsm.Get([]byte("key1"))
	if !found || string(val) != "value_that_is_quite_long_1" {
		t.Errorf("Data lost after flush!")
	}
    
    lsm.Close()
}
