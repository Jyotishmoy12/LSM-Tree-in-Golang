package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/engine/memtable"
	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/engine/sstable"
)

// The Engine’s job is to coordinate them: when you call Get, it first checks the MemTable,
// then searches through the SSTables on disk from newest to oldest.
// When you call Put and the MemTable fills up, the Engine triggers a flush.

// LSM represents the core database engine
type LSM struct {
	mu         sync.RWMutex
	memTable   *memtable.MemTable
	sstTables  []*sstable.Reader
	dir        string
	maxMemSize int
}

// New opens the LSM engine in the specified directory
func New(dir string, maxMemSize int) (*LSM, error) {
	// 0755 means the owner can read/write/execute, and others can read/execute
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	lsm := &LSM{
		dir:        dir,
		maxMemSize: maxMemSize,
	}
	// 1. Initialize MemTable
	// In a read DB we would replay the WAL here first to restore the MemTable state, but for simplicity we start fresh
	walPath := filepath.Join(dir, "active.wal")
	mt, err := memtable.NewMemTable(walPath, maxMemSize)
	if err != nil {
		return nil, err
	}
	lsm.memTable = mt
	// 2. Load existing SSTables
	if err := lsm.loadSSTables(); err != nil {
		return nil, err
	}
	return lsm, nil
}

// loadSSTables scans the directory for existing SSTable files and loads them into memory
func (lsm *LSM) loadSSTables() error {
	files, err := os.ReadDir(lsm.dir)
	if err != nil {
		return err
	}
	var sstFiles []string
	for _, f := range files {
		// .sst stands for "Sorted String Table", a common file format for LSM trees
		if strings.HasSuffix(f.Name(), ".sst") {
			sstFiles = append(sstFiles, filepath.Join(lsm.dir, f.Name()))
		}
	}
	// Sort files so newest (highest timestamp/index) are first
	sort.Sort(sort.Reverse(sort.StringSlice(sstFiles)))

	for _, path := range sstFiles {
		reader, err := sstable.Open(path)
		if err != nil {
			return err
		}
		lsm.sstTables = append(lsm.sstTables, reader)
	}
	return nil
}

// put adds a key-value pair to the MemTable, and flushes to disk if the MemTable is full.
func (lsm *LSM) Put(key, value []byte) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()
	if err := lsm.memTable.Put(key, value); err != nil {
		return err
	}
	// check if memTable is full and flush if needed
	if lsm.memTable.IsFull() {
		return lsm.flush()
	}
	return nil
}

// Get retrieves a value. It checks MemTable first and then searches through SSTables in order.
func (lsm *LSM) Get(key []byte) ([]byte, bool, error) {
	lsm.mu.RLock()
	defer lsm.mu.RUnlock()
	// 1. Check MemTable
	if val, found := lsm.memTable.Get(key); found {
		if string(val) == "TOMBSTONE_MARKER" {
			return nil, false, nil
		}
		return val, true, nil
	}
	// 2. Check SSTables
	for _, sst := range lsm.sstTables {
		val, found, err := sst.Get(key)
		if err != nil {
			return nil, false, err
		}
		if found {
			if string(val) == "TOMBSTONE_MARKER" {
				return nil, false, nil
			}
			return val, true, nil
		}
	}
	return nil, false, nil
}

// flush writes the current MemTable to a new SSTable on disk and resets the MemTable.
func (l *LSM) flush() error {
	// 1. Generate a unique filename based on timestamp
	sstPath := filepath.Join(l.dir, fmt.Sprintf("%d.sst", time.Now().UnixNano()))
	writer, err := sstable.NewWriter(sstPath)
	if err != nil {
		return err
	}

	// 2. Iterate over skiplist and write to SSTable
	it := l.memTable.GetIterator()
	for node := it; node != nil; node = node.Next() {
		if err := writer.WritePair(node.Key(), node.Value(), 0); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	// 3. Open the newly created sstable for reading
	reader, err := sstable.Open(sstPath)
	if err != nil {
		return err
	}

	// Prepend to readers (newest first)
	l.sstTables = append([]*sstable.Reader{reader}, l.sstTables...)

	// 4. Reset MemTable and WAL
	l.memTable.Close()
	os.Remove(filepath.Join(l.dir, "active.wal"))

	newWalPath := filepath.Join(l.dir, "active.wal")
	newMemTable, err := memtable.NewMemTable(newWalPath, l.maxMemSize)
	if err != nil {
		return err
	}
	l.memTable = newMemTable
	return nil
}

func (l *LSM) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.memTable.Close(); err != nil {
		return err
	}

	for _, sst := range l.sstTables {
		if err := sst.Close(); err != nil {
			return err
		}
	}
	return nil
}

// compact merges two oldeest SSTables into a single one to reduce read amplification.
// This is a simplified version that just merges the two oldest SSTables without any optimization.

// Compact merges the two oldest SSTables into a single one to reduce read amplification.
func (l *LSM) Compact() error {
	l.mu.Lock()
	if len(l.sstTables) < 2 {
		l.mu.Unlock()
		return nil // Nothing to compact
	}

	// For simplicity, we merge the two oldest (last two in our slice)
	// In a production DB, this would happen in a background goroutine
	t1 := l.sstTables[len(l.sstTables)-2]
	t2 := l.sstTables[len(l.sstTables)-1]
	l.mu.Unlock()

	// 1. Create a new SSTable for the merged data
	compactedPath := filepath.Join(l.dir, fmt.Sprintf("compacted_%d.sst", time.Now().UnixNano()))
	writer, err := sstable.NewWriter(compactedPath)
	if err != nil {
		return err
	}

	// 2. Perform a Merge Sort between the two SSTables
	// Since we don't have iterators for SSTables yet, we'll use a simplified
	// approach: Load keys and merge. (In a real DB, we stream them).

	// Implementation Note: To keep this concise, we will use a map to de-duplicate
	// but in God Mode, you'd use a priority queue for streaming merge.
	mergedData := make(map[string][]byte)

	// We iterate through both, newer data from t1 (if exists) overrides t2
	// But since t1 is newer than t2 in our slice logic:
	l.loadIntoMap(t2, mergedData)
	l.loadIntoMap(t1, mergedData)

	// Sort keys to maintain SSTable contract
	keys := make([]string, 0, len(mergedData))
	for k := range mergedData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		writer.WritePair([]byte(k), mergedData[k], 0)
	}
	writer.Close()

	// 3. Update the Engine State
	l.mu.Lock()
	defer l.mu.Unlock()

	// Open the new one
	newReader, _ := sstable.Open(compactedPath)

	// Remove the two old ones and add the new one
	l.sstTables = append(l.sstTables[:len(l.sstTables)-2], newReader)

	return nil
}

// Helper to load SSTable data into a map for merging
func (l *LSM) loadIntoMap(r *sstable.Reader, data map[string][]byte) {
	// In Phase 4, we built the index. We can use it to iterate.
	for _, entry := range r.GetIndex() {
		val, found, _ := r.Get(entry.Key)
		if found {
			data[string(entry.Key)] = val
		}
	}
}

// Delete inserts a tombstone for the given key.
func (l *LSM) Delete(key []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// A delete is just a Put with a special "Tombstone" value or flag.
	// For simplicity, we'll use a reserved string or a separate field.
	return l.memTable.Put(key, []byte("TOMBSTONE_MARKER"))
}
