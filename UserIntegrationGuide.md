# Go-LSM: User Deployment & Integration Guide

Go-LSM is a persistent, write-optimized Key-Value storage engine built on the Log-Structured Merge-Tree architecture. It is designed for high-ingestion workloads and offers both direct Go library integration and a standalone TCP server.

---

## 1. Deployment Options

### Option A: Using Docker (Recommended)

The fastest way to get Go-LSM running in a production-like environment without installing Go.

**Build the Image:**

```bash
docker build -t go-lsm-db .
```

**Run with Persistent Storage:**

This maps port 6379 and ensures your data survives if the container stops.

```bash
docker run -p 6379:6379 -v ${PWD}/stress_storage:/root/stress_storage go-lsm-db
```

### Option B: Manual Setup (Local)

Use this if you want to run the server directly on your machine.

**Clone and Build:**

```bash
git clone https://github.com/Jyotishmoy12/LSM-Tree-in-Golang
cd go-lsm
go build -o lsm-server ./cmd/lsm-server/main.go
```

**Start the Server:**

```bash
./lsm-server
```

The server will start listening on port 6379.

---

## 2. How to Connect & Use

Once the server is running, you can interact with it using various tools.

### Via Terminal (using Ncat/Netcat)

If you have Nmap installed, use `ncat`. Otherwise, use `nc`.

```bash
ncat localhost 6379
```

### Available Commands

- **SET <key> <value>:** Store data.
- **GET <key>:** Retrieve data.
- **DELETE <key>:** Mark a key for removal.
- **COMPACT:** Merge SSTable files to optimize performance.
- **QUIT:** Close the connection.

---

## 3. Programming Language Integration

You can talk to Go-LSM from any language that supports TCP Sockets.

### Python Integration

```python
import socket

def send_command(cmd):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.connect(("localhost", 6379))
        s.sendall(f"{cmd}\n".encode())
        return s.recv(1024).decode()

# Examples
print(send_command("SET username jyotishmoy")) # Returns: OK
print(send_command("GET username"))            # Returns: "jyotishmoy"
```

### Node.js Integration

```javascript
const net = require('net');

const client = net.createConnection({ port: 6379 }, () => {
    client.write('SET platform "LSM-Tree"\n');
});

client.on('data', (data) => {
    console.log('Response:', data.toString());
    client.end();
});
```
### Golang Integration

## Installation: 

```
go get go get github.com/Jyotishmoy12/LSM-Tree-in-Golang

```

```Golang
package main

import (
	"fmt"
	"log"
	"github.com/Jyotishmoy12/LSM-Tree-in-Golang/engine"
)

func main() {
	// 1. Initialize the engine
	// We point it to a local folder and set a 1KB MemTable limit
	db, err := engine.New("./my_db_data", 1024)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer db.Close()

	// 2. Put some data
	fmt.Println("Writing data...")
	err = db.Put([]byte("hero"), []byte("Batman"))
	if err != nil {
		fmt.Printf("Put error: %v\n", err)
	}

	// 3. Retrieve data
	val, found, err := db.Get([]byte("hero"))
	if err != nil {
		fmt.Printf("Get error: %v\n", err)
	}

	if found {
		fmt.Printf(" Found: %s\n", string(val))
	} else {
		fmt.Println(" Key not found")
	}
}
```

---

## 4. Inspection & Debugging Tools

Go-LSM comes with built-in utilities to "see" inside the database files.

| Tool | Purpose | Command |
|------|---------|---------|
| SSTable Dump | View sorted disk data | `go run ./cmd/lsm-dump ./stress_storage/<file>.sst` |
| WAL Dump | View unflushed recovery logs | `go run ./cmd/lsm-wal-dump ./stress_storage/active.wal` |
| Stress Test | Auto-generate 100+ writes | `go run ./cmd/lsm-stress` |

---

## 5. Using as a Go Library

If you are a Go developer, you don't need the server; you can embed the engine directly.

```go
import "github.com/jyotishmoy12/go-lsm/internal"

func main() {
    // Initialize engine with 1MB MemTable limit
    db, _ := internal.New("./my_storage", 1024*1024) 
    defer db.Close()

    db.Put([]byte("hero"), []byte("batman"))
}
```

---

## Final Notes for Users

- **Case Sensitivity:** Keys are case-sensitive (e.g., `User` and `user` are different).
- **Persistence:** All data is written to the Write-Ahead Log (WAL) immediately, ensuring it survives a crash.
- **Storage:** Default data is stored in the `./stress_storage` directory unless configured otherwise.
