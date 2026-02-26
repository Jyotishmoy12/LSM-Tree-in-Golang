package engine

import (
	"os"
	"testing"
)

func TestLSM_Compaction(t *testing.T) {
	dir := "compaction_test"
	defer os.RemoveAll(dir)

	lsm, _ := New(dir, 100)

	// 1. Force two flushes by writing data
	lsm.Put([]byte("a"), []byte("1"))
	lsm.flush() // Manual flush for testing
	lsm.Put([]byte("b"), []byte("2"))
	lsm.flush()

	if len(lsm.sstTables) != 2 {
		t.Fatalf("Expected 2 SSTables, got %d", len(lsm.sstTables))
	}

	// 2. Run Compaction
	err := lsm.Compact()
	if err != nil {
		t.Fatalf("Compaction failed: %v", err)
	}

	// 3. Verify
	if len(lsm.sstTables) != 1 {
		t.Errorf("Expected 1 SSTable after compaction, got %d", len(lsm.sstTables))
	}

	val, _, _ := lsm.Get([]byte("a"))
	if string(val) != "1" {
		t.Errorf("Data lost during compaction")
	}
}
