package sstable

import (
	"os"
	"testing"
)

func TestSSTable_Write(t *testing.T) {
	path := "test.sst"
	defer os.Remove(path)

	writer, err := NewWriter(path)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	pairs := []struct {
		k, v string
	}{
		{"key1", "val1"},
		{"key2", "val2"},
		{"key3", "val3"},
	}

	for _, p := range pairs {
		if err := writer.WritePair([]byte(p.k), []byte(p.v), 0); err != nil {
			t.Errorf("Failed to write pair: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Check if file exists
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		t.Error("SSTable file was not created")
	}
}
