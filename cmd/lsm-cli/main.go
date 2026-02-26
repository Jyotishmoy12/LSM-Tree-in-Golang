package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/engine"
)

func main() {
	// Initialize the LSM engine with 1KB MemTable limit for easy testing of flushes
	db, err := engine.New("./stress_storage", 1024)
	if err != nil {
		fmt.Printf("Failed to initialize DB: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("LSM-Tree initialized.")
	fmt.Println("Commands: SET <key> <val> | GET <key> | COMPACT | EXIT")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])

		switch command {
		case "SET":
			if len(parts) < 3 {
				fmt.Println("Usage: SET <key> <val>")
				continue
			}
			key, val := parts[1], parts[2]
			err := db.Put([]byte(key), []byte(val))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "GET":
			if len(parts) < 2 {
				fmt.Println("Usage: GET <key>")
				continue
			}
			key := parts[1]
			val, found, err := db.Get([]byte(key))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if !found {
				fmt.Println("(nil)")
			} else {
				fmt.Printf("\"%s\"\n", string(val))
			}
		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: DELETE <key>")
				continue
			}
			err := db.Delete([]byte(parts[1]))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK (Tombstone added)")
			}

		case "COMPACT":
			fmt.Println("Starting compaction...")
			err := db.Compact()
			if err != nil {
				fmt.Printf("Compaction error: %v\n", err)
			} else {
				fmt.Println("Compaction complete.")
			}

		case "EXIT":
			fmt.Println("Shutting down...")
			return

		default:
			fmt.Println("Unknown command. Try SET, GET, COMPACT, or EXIT.")
		}
	}
}
