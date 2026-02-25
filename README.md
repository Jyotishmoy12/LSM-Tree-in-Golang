# Go-LSM: A Production-Grade LSM-Tree Implementation

Go-LSM is a high-performance, persistent Key-Value storage engine written in Go. It implements the Log-Structured Merge-Tree architecture—the industry standard for write-heavy workloads—providing durability, sorted persistence, and high-speed data retrieval.

---

## The Technical Stack

We built a multi-layered storage engine that manages data from volatile RAM to immutable disk storage. The engine is divided into five core subsystems:

### 1. The Durability Layer (Write-Ahead Log)

To ensure no data is lost during a crash, we implemented a WAL.

- **Sequential I/O:** Every write is appended to the WAL using `os.O_APPEND` to maximize disk throughput.
- **Binary Encoding:** Data is stored in a `[Type][KeyLen][ValLen][Key][Value]` format using `binary.LittleEndian` for cross-platform portability.
- **Hardware Sync:** We use `file.Sync()` to force the OS kernel to flush buffers to physical storage, ensuring absolute durability.

### 2. The In-Memory Layer (SkipList MemTable)

Since disk is slow, we buffer writes in a SkipList.

- **Probabilistic Balancing:** The SkipList provides O(log n) search and insertion without the complex rebalancing logic of Red-Black trees.
- **Sorted Order:** The SkipList ensures that data is always sorted in RAM, which is the prerequisite for creating SSTables.
- **Threshold Management:** Once the MemTable reaches its size limit (e.g., 512 bytes in our stress test), it triggers an automatic "flush" to disk.

### 3. The Persistence Layer (SSTables)

SSTables (Sorted String Tables) are the heart of the LSM-Tree's storage.

- **Footer-Based Indexing:** Every SSTable ends with a Footer (the last 8 bytes) that points to an Index Block. This allows the engine to jump straight to the index without scanning the file.
- **Binary Search:** Because keys are sorted, we perform a binary search on the in-memory index to find the exact byte offset of any key on disk.

### 4. The Maintenance Layer (Compaction)

To prevent "Read Amplification" (checking too many files), we built a Compaction Engine.

- **K-Way Merge:** We merge multiple sorted files into one, similar to the merge phase of Merge Sort.
- **Tombstone Processing:** Deletions are handled via "Tombstones." During compaction, the engine identifies these markers and permanently removes the deleted data from disk.

### 5. The Tooling Suite

- **lsm-cli:** A REPL for manual database interaction.
- **lsm-stress:** An automated load tester to verify engine stability under pressure.
- **lsm-dump & lsm-wal-dump:** Custom binary parsers that transform raw bytes into human-readable tables.

---

## How We Built It: Step-by-Step Evolution

### Phase 1: Foundation & Durability

We started by creating the WAL. We defined the binary protocol and ensured that every write was synced to the hardware before being acknowledged.

### Phase 2: Sorted RAM Storage

We implemented the SkipList from scratch. We focused on pointer-based level management and lexicographical byte comparison to keep keys sorted.

### Phase 3: The Persistence Bridge

We built the SSTable Writer. This involved creating an Iterator for the SkipList that could "walk" the memory in order and stream that data into a binary file. We implemented the Footer logic here, ensuring the last 8 bytes of every file contain the pointer to the index.

### Phase 4: The Read Path

We built the SSTable Reader. We implemented a "tail-first" reading strategy:

1. Seek to End - 8
2. Read the Index Offset
3. Seek to Offset and load the Index
4. Execute Binary Search

### Phase 5: The Orchestrator

We created the LSM struct. This "Brain" coordinates the search: first checking the MemTable, then checking the SSTable layers from newest to oldest.

---

## Sample Output & Verification

### SSTable Dump (Sorted Persistence)

When you run our dump tool, you see the perfectly ordered results of our SkipList-to-Disk flush:

```
--- Dumping SSTable: ./stress_storage/1772...sst ---
KEY                  | VALUE
---------------------------------------------
key-000              | value-data-block-000...
key-001              | value-data-block-001...
```

### WAL Dump (Crash Recovery Log)

The WAL dump shows the "unflushed" writes currently waiting in the buffer:

```
--- Dumping WAL: ./stress_storage/active.wal ---
key-099              | value-data-block-099...
```

---

## Summary

This project demonstrates a full-stack understanding of database internals, from low-level binary manipulation and file I/O to high-level concurrency management and data structure optimization.

For detailed testing instructions, see [TestingGuide.md](TestingGuide.md).