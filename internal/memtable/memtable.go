package memtable

import (
	"fmt"

	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/internal/wal"
)

// MemTable is an in-memory data structure that holds key-value pairs before they are flushed to disk.
// When the MemTable reaches a certain size (e.g., 4MB), it becomes "immutable" and a new one is created.

// memtable coordinates the SkipList and the Write-Ahead Log (WAL) to ensure data durability and efficient in-memory operations.
type MemTable struct {
	list     *SkipList
	wal      *wal.WAL
	maxSize  int
	currSize int
}

// newMemTable initializes a new memTable with the given WAL and maximum size.
func NewMemTable(walPath string, maxSize int) (*MemTable, error) {
	w, err := wal.New(walPath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize WAL: %w", err)
	}
	return &MemTable{
		list:    NewSkipList(),
		wal:     w,
		maxSize: maxSize,
	}, nil
}

// Put inserts a key-value pair into the memTable. It first writes to the WAL for durability, then updates the SkipList.
func (m *MemTable) Put(key, value []byte) error {
	// 1. Write to WAL
	if err := m.wal.Write(key, value); err != nil {
		return fmt.Errorf("failed to write to WAL: %w", err)
	}
	// 2. Update SkipList
	m.list.Put(key, value)

	// 3. Track size ( simplified: key len + value len )
	m.currSize += len(key) + len(value)
	return nil
}

// Get reads from the SkipList. If the key is not found, it returns nil.
func (m *MemTable) Get(key []byte) ([]byte, bool) {
	return m.list.Get(key)
}

// IsFull checks if the memTable has reached its maximum size.
func (m *MemTable) IsFull() bool {
	return m.currSize >= m.maxSize
}

// Close closes the WAL file.
func (m *MemTable) Close() error {
	return m.wal.Close()
}

func (m *MemTable) GetIterator() *Node {
	return m.list.head.next[0]
}
