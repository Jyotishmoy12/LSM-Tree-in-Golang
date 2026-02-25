package sstable

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// SSTable (Sorted String Table) is a file format used in LSM-trees to store sorted key-value pairs on disk.
// It’s designed for efficient reads and writes, especially for large datasets that exceed memory limits.

// SSTable isn't just a list of data. It’s split into:

// Data Blocks: The actual K-V pairs.

// Index Block: Offsets telling us where specific keys are (so we don't scan the whole file).

// Footer: Metadata and the Bloom Filter.

// IndexEntry holds the location of a key in the data file.
type IndexEntry struct {
	Key    []byte
	Offset int64
}

// Writer handles the creation of a new SSTable file.
type Writer struct {
	file  *os.File
	index []IndexEntry
}

// NewWriter initializes a writer for a specific file path.
func NewWriter(path string) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSTable: %w", err)
	}
	return &Writer{file: f}, nil
}

// WritePair appends a K-V pair to the data section and tracks its index.
func (w *Writer) WritePair(key, value []byte, entryType byte) error {
	offset, err := w.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	// Record the index entry
	w.index = append(w.index, IndexEntry{Key: key, Offset: offset})

	// Binary Format: [KeyLen(4)][ValLen(4)][Key][Value]
	buf := make([]byte, 9)
	buf[0] = entryType
	binary.LittleEndian.PutUint32(buf[1:5], uint32(len(key)))
	binary.LittleEndian.PutUint32(buf[5:9], uint32(len(value)))

	if _, err := w.file.Write(buf); err != nil {
		return err
	}
	if _, err := w.file.Write(key); err != nil {
		return err
	}
	if _, err := w.file.Write(value); err != nil {
		return err
	}

	return nil
}

// Close finalizing the SSTable by writing the Index and Footer.
func (w *Writer) Close() error {
	// 1. Record where the Index starts
	indexOffset, _ := w.file.Seek(0, io.SeekCurrent)

	// 2. Write the Index entries
	for _, entry := range w.index {
		buf := make([]byte, 12) // 4 for KeyLen, 8 for Offset
		binary.LittleEndian.PutUint32(buf[0:4], uint32(len(entry.Key)))
		binary.LittleEndian.PutUint64(buf[4:12], uint64(entry.Offset))
		w.file.Write(buf)
		w.file.Write(entry.Key)
	}

	// 3. Write Footer: [IndexOffset (8 bytes)]
	footer := make([]byte, 8)
	binary.LittleEndian.PutUint64(footer, uint64(indexOffset))
	w.file.Write(footer)

	return w.file.Close()
}
