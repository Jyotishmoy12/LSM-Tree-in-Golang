package wal

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WAL (Write-Ahead Log) is a technique used in databases and file systems to ensure data integrity and durability.
// It works by recording changes to a log before they are applied to the main data store.
// This allows for recovery in case of crashes or failures, as the log can be replayed to restore the system to a consistent state.

type WAL struct {
	file *os.File
}

// new creates a new WAL file or opens an existing one.

func New(path string) (*WAL, error) {
	// O_APPEND: Append only for high-speed sequential writes.
	// O_CREATE: Create if not exists.
	// O_RDWR: Read/Write access.
	// 0644: File permissions (owner read/write, group read, others read).

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file: %w", err)
	}
	return &WAL{file: f}, nil
}

// Write appends a log entry to the WAL file.
// Format: [KeyLen (4 bytes)][ValueLen (4 bytes)][Key Content][Value Content]
func (w *WAL) Write(key, value []byte) error {
	// create an 8-byte buffer for the two 4-byte uint32 lengths
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(len(key)))   // Key length
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(value))) // Value length

	// Write the header (lengths) followed by the key and value content
	if _, err := w.file.Write(header); err != nil {
		return fmt.Errorf("failed to write header to WAL: %w", err)
	}
	// Write the key content
	if _, err := w.file.Write(key); err != nil {
		return fmt.Errorf("failed to write key to WAL: %w", err)
	}
	// Write the value content
	if _, err := w.file.Write(value); err != nil {
		return fmt.Errorf("failed to write value to WAL: %w", err)
	}
	return w.file.Sync() // Ensure data is flushed to disk
}

// Close closes the WAL file.
func (w *WAL) Close() error {
	return w.file.Close()
}
