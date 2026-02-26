package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/engine/sstable"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/lsm-dump <path-to-sstable>")
		return
	}

	path := os.Args[1]
	reader, err := sstable.Open(path)
	if err != nil {
		log.Fatalf("Failed to open SSTable: %v", err)
	}
	defer reader.Close()

	fmt.Printf("--- Dumping SSTable: %s ---\n", path)
	fmt.Printf("%-20s | %-20s\n", "KEY", "VALUE")
	fmt.Println(strings.Repeat("-", 45))

	// We use the GetIndex method we exported in Phase 6 to iterate
	for _, entry := range reader.GetIndex() {
		val, found, err := reader.Get(entry.Key)
		if err != nil {
			fmt.Printf("Error reading key %s: %v\n", string(entry.Key), err)
			continue
		}
		if found {
			fmt.Printf("%-20s | %-20s\n", string(entry.Key), string(val))
		}
	}
	fmt.Println("--- End of Dump ---")
}
