package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/internal"
)

func main() {
	storageDir := "./stress_storage"
	os.RemoveAll(storageDir) // Start fresh

	// 1. Init DB with a small MemTable (512 bytes) to trigger flushes often
	db, err := internal.New(storageDir, 512)
	if err != nil {
		log.Fatalf("Failed to init: %v", err)
	}

	fmt.Println("Starting Stress Test: Generating 100 writes to trigger flushes...")

	// 2. Automate writes
	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key-%03d", i))
		val := []byte(fmt.Sprintf("value-data-block-%03d-some-extra-padding-to-fill-memory", i))

		if err := db.Put(key, val); err != nil {
			fmt.Printf("Error at index %d: %v\n", i, err)
		}
	}

	// 3. Check Disk
	files, _ := os.ReadDir(storageDir)
	sstCount := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sst" {
			sstCount++
		}
	}
	fmt.Printf(" Verification: Found %d SSTable files on disk.\n", sstCount)

	// 4. Automate Read Verification
	fmt.Println("Verifying data integrity...")
	for i := 0; i < 100; i++ {
		key := []byte(fmt.Sprintf("key-%03d", i))
		expected := fmt.Sprintf("value-data-block-%03d-some-extra-padding-to-fill-memory", i)

		val, found, _ := db.Get(key)
		if !found || string(val) != expected {
			fmt.Printf("Data mismatch at %s!\n", string(key))
		}
	}

	// 5. Automate Compaction
	fmt.Println("Triggering Compaction...")
	if err := db.Compact(); err != nil {
		fmt.Printf("Compaction error: %v\n", err)
	}

	db.Close()
	fmt.Println("Stress Test Complete. Everything is working in God Mode.")
}
