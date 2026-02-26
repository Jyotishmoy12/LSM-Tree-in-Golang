package wal

import (
	"os"
	"testing"
)

func TestWAL_Write(t *testing.T) {
	tempPath := "test.log"
	defer os.Remove(tempPath) // Clean up after test

	w, err := New(tempPath)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}

	key := []byte("username")
	val := []byte("jyotishmoy")

	err = w.Write(key, val)
	if err != nil {
		t.Errorf("Failed to write to WAL: %v", err)
	}

	// Verify file exists and has data
	info, err := os.Stat(tempPath)
	if err != nil {
		t.Fatalf("File info error: %v", err)
	}

	// Expected size: 8 (header) + 8 (key) + 10 (val) = 26 bytes
	expected := int64(8 + len(key) + len(val))
	if info.Size() != expected {
		t.Errorf("Expected size %d, got %d", expected, info.Size())
	}

	w.Close()
}
