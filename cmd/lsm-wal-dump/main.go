package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/lsm-wal-dump <path-to-wal>")
		return
	}

	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open WAL: %v", err)
	}
	defer file.Close()

	fmt.Printf("--- Dumping WAL: %s ---\n", path)
	fmt.Printf("%-20s | %-20s\n", "KEY", "VALUE")
	fmt.Println(strings.Repeat("-", 45))

	for {
		// 1. Read the 8-byte header [KeyLen(4)][ValLen(4)]
		header := make([]byte, 8)
		_, err := io.ReadFull(file, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading header: %v\n", err)
			break
		}

		keyLen := binary.LittleEndian.Uint32(header[0:4])
		valLen := binary.LittleEndian.Uint32(header[4:8])

		// 2. Read the Key
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(file, key); err != nil {
			fmt.Printf("Error reading key: %v\n", err)
			break
		}

		// 3. Read the Value
		value := make([]byte, valLen)
		if _, err := io.ReadFull(file, value); err != nil {
			fmt.Printf("Error reading value: %v\n", err)
			break
		}

		fmt.Printf("%-20s | %-20s\n", string(key), string(value))
	}
	fmt.Println("--- End of WAL Dump ---")
}