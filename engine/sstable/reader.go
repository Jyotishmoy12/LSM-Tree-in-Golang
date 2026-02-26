package sstable

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// reader helps to read sstable file,
// it is used by sstable iterator and sstable merger.

// The reading process is as follows:

// Jump to the last 8 bytes of the file to find where the Index starts.

// Load only the Index into memory.

// Use Binary Search on the index to find the exact byte offset of the data.

// Jump to that offset and read the value.

// Reader allows for efficient reading of an SSTable file.
type Reader struct {
	file  *os.File
	index []IndexEntry
}

// Open loads an SSTable file and prepares it for reading.
func Open(filePath string) (*Reader, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSTable: %w", err)
	}
	r := &Reader{file: f}
	if err := r.loadIndex(); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to load index: %w", err)
	}
	return r, nil
}

// loadIndex reads the index from the SSTable file and stores it in memory.
func (r *Reader) loadIndex() error {
	// 1. Seek to the last 8 bytes (the footer)
	_, err := r.file.Seek(-8, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to footer: %w", err)
	}

	// 2. Read index offset
	var indexOffset uint64
	if err := binary.Read(r.file, binary.LittleEndian, &indexOffset); err != nil {
		return fmt.Errorf("failed to read footer: %w", err)
	}

	// 3. Seek to the index start
	_, err = r.file.Seek(int64(indexOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to index: %w", err)
	}

	// 4. Read the index entries until the end of file (minus footer)
	for {
		var keyLen uint32
		err := binary.Read(r.file, binary.LittleEndian, &keyLen)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		// If we hit the footer (8 bytes left), stop
		currPos, _ := r.file.Seek(0, io.SeekCurrent)
		fileInfo, _ := r.file.Stat()
		if currPos >= fileInfo.Size()-8 {
			break
		}

		var offset uint64
		binary.Read(r.file, binary.LittleEndian, &offset)

		key := make([]byte, keyLen)
		r.file.Read(key)

		r.index = append(r.index, IndexEntry{Key: key, Offset: int64(offset)})
	}
	return nil
}

// Get retrieves the value associated with the given key using binary search on the index.
// becoz sstable is sorted so binary search is very efficient
func (r *Reader) Get(key []byte) ([]byte, bool, error) {
	// Binary search on the index
	low, high := 0, len(r.index)-1
	var foundEntry *IndexEntry
	for low <= high {
		mid := low + (high-low)/2
		cmp := bytes.Compare(r.index[mid].Key, key)
		if cmp == 0 {
			foundEntry = &r.index[mid]
			break
		} else if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	if foundEntry == nil {
		return nil, false, nil // Key not found
	}
	// Seek to the data block offset
	_, err := r.file.Seek(foundEntry.Offset, io.SeekStart)
	if err != nil {
		return nil, false, fmt.Errorf("failed to seek to data block: %w", err)
	}
	//Read header: [keyLen(4)][valueLen(4)] // headers: They are known as the first 8 bytes of the data block, which contain the lengths of the key and value. This allows us to know how many bytes to read for the key and value, respectively.
	header := make([]byte, 9) // 8 bytes for lengths + 1 byte for potential future use
	if _, err := r.file.Read(header); err != nil {
		return nil, false, fmt.Errorf("failed to read data block header: %w", err)
	}
	entryType := header[0]
	keyLen := binary.LittleEndian.Uint32(header[1:5])
	valueLen := binary.LittleEndian.Uint32(header[5:9])

	if entryType == 1 {
		return nil, false, nil // tombstone entry, key is deleted
	}

	// Skip the key (we already know it) and read the value
	if _, err := r.file.Seek(int64(keyLen), io.SeekCurrent); err != nil {
		return nil, false, fmt.Errorf("failed to skip key: %w", err)
	}
	value := make([]byte, valueLen)
	if _, err := r.file.Read(value); err != nil {
		return nil, false, fmt.Errorf("failed to read value: %w", err)
	}
	return value, true, nil
}

// Close releases any resources held by the Reader.
func (r *Reader) Close() error {
	return r.file.Close()
}

// GetIndex returns the index entries for compaction/iteration.
func (r *Reader) GetIndex() []IndexEntry {
	return r.index
}
