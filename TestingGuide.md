# LSM-Tree User Testing & Inspection Guide

This guide explains how to utilize the four custom tools you've built to verify data durability, sorted persistence, and storage optimization.

---

## 1. The Ingestion Phase (Stress Testing)

Before you can inspect data, you need to generate it.

**Action:** Run the automated stress tester.

```bash
go run ./cmd/lsm-stress
```

**What happens:** This script bypasses the CLI and floods the engine with 100 high-volume writes. Because the MemTable limit is set low (512 bytes), you will witness the engine automatically "flushing" data to disk.

**Result:** A new folder `./stress_storage` will appear containing approximately 11 `.sst` files and one `active.wal`.

---

## 2. The Inspection Phase (Internal Dumps)

Standard text editors cannot read your database files because they are stored in a custom binary format. Use the following dumpers:

### 2.1 Inside the WAL (Write-Ahead Log)

The WAL contains the "volatile" data—writes that occurred but haven't been turned into an SSTable yet.

```bash
go run ./cmd/lsm-wal-dump ./stress_storage/active.wal
```

**Human-Readable Output:**

```
KEY                  | VALUE
---------------------------------------------
key-099              | value-data-block-099...
```

**Insight:** If the system crashes, this file is what the engine reads to restore the MemTable.

### 2.2 Inside the SSTables (Sorted String Tables)

These are your immutable, sorted data files.

```bash
go run ./cmd/lsm-dump ./stress_storage/<timestamp>.sst
```

**Human-Readable Output:**

```
KEY                  | VALUE
---------------------------------------------
key-000              | value-data-block-000...
key-001              | value-data-block-001...
```

**Insight:** Notice how keys are in perfect alphabetical order. This allows the engine to use Binary Search to find any value in O(log n) time.

---

## 3. The Interactive Phase (Manual CLI)

Now, step into the role of the database administrator.

```bash
go run ./cmd/lsm-cli
```

> **Note:** Ensure it points to `./stress_storage`.

### Testing Logic

- **Read:** `GET key-050` — Confirms the engine can search across multiple disk layers.
- **Shadow:** `SET key-050 "NewValue"` — If you GET it again, the engine returns `"NewValue"` because the MemTable shadows the old disk data.
- **Delete:** `DELETE key-050` — This writes a Tombstone. The data is still on disk, but the engine will now return `(nil)`.

---

## 4. The Maintenance Phase (Compaction)

Eventually, having 11 files makes reads slow. Compaction fixes this.

**Action:** Inside the CLI, type `COMPACT`.

**What happens:** The engine performs a K-Way Merge Sort. It combines the multiple `.sst` files into one, removes the old version of `key-050`, and officially deletes anything marked with a tombstone.

**Verification:** Run `ls ./stress_storage`. You will see the many small files replaced by a single `compacted_...sst` file.

---

## Storage Format Breakdown

| File Type | Storage Logic | Content Structure |
|-----------|---------------|-------------------|
| `.wal` | Append-only log | `[Type][KeyLen][ValLen][Key][Value]` |
| `.sst` | Sorted, Indexed blocks | `[Data Blocks] + [Index Block] + [Footer]` |

### Example SSTable Dump Result

```
--- Dumping SSTable: ./stress_storage/17720344...sst ---
key-000 : value-data-block-000...
key-001 : value-data-block-001...
```

### Example WAL Dump Result

```
--- Dumping WAL: ./stress_storage/active.wal ---
key-099 : value-data-block-099...
```
