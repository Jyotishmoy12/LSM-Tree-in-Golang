package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/jyotishmoy12/go-lsm/internal"
)

func main() {
	// 1. Initialize the Engine
	db, err := internal.New("./stress_storage", 1024*1024)
	if err != nil {
		fmt.Printf("Failed to start engine: %v\n", err)
		return
	}
	defer db.Close()

	// 2. Start TCP Listener on port 6379
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Printf("Failed to bind port: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("LSM-Server listening on :6379")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Connection error: %v\n", err)
			continue
		}
		// Handle each client in a separate goroutine for high performance
		go handleConnection(conn, db)
	}
}

func handleConnection(conn net.Conn, db *internal.LSM) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		command := strings.ToUpper(parts[0])
		var response string

		switch command {
		case "SET":
			if len(parts) < 3 {
				response = "ERR usage: SET <key> <val>\n"
			} else {
				err := db.Put([]byte(parts[1]), []byte(parts[2]))
				if err != nil {
					response = fmt.Sprintf("ERR %v\n", err)
				} else {
					response = "OK\n"
				}
			}
		case "GET":
			if len(parts) < 2 {
				response = "ERR usage: GET <key>\n"
			} else {
				val, found, _ := db.Get([]byte(parts[1]))
				if !found {
					response = "(nil)\n"
				} else {
					response = fmt.Sprintf("\"%s\"\n", string(val))
				}
			}
		case "QUIT":
			conn.Write([]byte("BYE\n"))
			return
		default:
			response = "ERR unknown command\n"
		}
		conn.Write([]byte(response))
	}
}
