# Google File System (GFS) Project in Golang

Building a Google File System–style project in Go is a great choice for you since you're already working with distributed systems ideas (Kafka, microservices, etc.). Below is a realistic step-by-step project plan that mirrors the architecture described in the The Google File System.

## Step-by-Step Guide

### Core Components

1. **Master**
   - Maintains metadata
   - Knows where chunks are stored
   - Handles namespace operations

2. **Chunk Servers**
   - Store actual data
   - Replicate chunks
   - Serve reads/writes

3. **Client**
   - Talks to master for metadata
   - Talks directly to chunk servers for data

### Important Concept
Files are split into 64MB chunks (in real GFS).

## Phase 1 — Basic Distributed File System (Single Machine Simulation)

**Goal:** Understand the mechanics first.

**Features:**
- File upload
- File download
- Chunk splitting
- Chunk mapping

**Project Structure:**
```
gfs/
├── master/
│   ├── master.go
│   └── metadata.go
├── chunkserver/
│   ├── server.go
│   └── storage.go
├── client/
│   └── client.go
├── proto/
│   └── gfs.proto
└── common/
    └── types.go
```

## Phase 2 — Implement the Master

**Responsibilities:**
- Maintain file namespace
- Map file → chunks
- Track chunk locations
- Handle replication decisions

**Metadata Example:**
```go
type ChunkInfo struct {
    ChunkID   string
    Locations []string
}

type FileMetadata struct {
    FileName string
    Chunks   []ChunkInfo
}
```

Master keeps: filename -> chunks -> servers

Store it in memory first. Later: persist to disk, add recovery.

## Phase 3 — Chunk Server

Each chunk server:
- Stores chunks on disk
- Replicates chunks
- Serves read/write

**Storage Layout:**
```
data/
├── chunk_18273
├── chunk_18274
└── chunk_18275
```

**Write Flow:**
- Client → Master
- Master returns chunk servers
- Client → Primary chunkserver
- Primary replicates to secondaries

## Phase 4 — gRPC Communication

Use gRPC between services.

- Client → Master
- Client → ChunkServer
- ChunkServer → ChunkServer

**Proto Example:**
```proto
service MasterService {
  rpc GetChunkLocations(GetChunkRequest) returns (ChunkLocationResponse);
}

service ChunkService {
  rpc WriteChunk(WriteRequest) returns (WriteResponse);
  rpc ReadChunk(ReadRequest) returns (ReadResponse);
}
```

Since you like Go backend projects, this fits perfectly.

## Phase 5 — File Write Flow (Real GFS Behavior)

**Steps:**
- Client asks master
- Master returns chunk servers
- Client pushes data to all replicas
- Primary orders write
- Secondaries apply write

**Pseudo Flow:**
- client -> master: where to write
- master -> client: chunkserver list
- client -> all chunkservers: send data
- client -> primary: commit
- primary -> secondary: replicate

## Phase 6 — Chunk Replication

Add replication factor = 3

Master decides placement.

**Example:** chunk_1 -> [server1, server3, server5]

Heartbeat from chunkservers: every 10 sec -> master

If server dies → re-replicate.

## Phase 7 — Fault Tolerance

Add:
- Heartbeats
- Chunkserver → Master: status, disk usage, chunk list
- Master detects failures
- If server missing: re-replicate chunks

## Phase 8 — Persistent Metadata

Right now master memory = dangerous.

Add:
- Write Ahead Log
- Operation log

**Example:**
```
CREATE file1
ADD_CHUNK chunk123
```

Recovery: load log, rebuild state

## Phase 9 — Snapshot Support

Feature from real GFS.

Copy-on-write snapshot.

snapshot(fileA)

Instead of copying: share chunks, duplicate on write

## Phase 10 — Advanced Improvements

1. **Lease System**
   - Master gives primary lease to chunkserver
   - Lease = 60s
   - Reduces master involvement

2. **Garbage Collection**
   - Deleted chunks cleaned later
   - Background worker

3. **Rebalancing**
   - Master moves chunks to balance storage

## Phase 11 — Deployment

Run using Docker.

**Example:**
- master: 1
- chunkservers: 3+
- client: CLI

Simulate cluster locally.

## Phase 12 — Observability

Add:
- Prometheus metrics
- Logging
- Request tracing

**Metrics:**
- Chunk count
- Replication lag
- Failed writes

## Phase 13 — Stretch Goals (Very Impressive)

- **Web UI**
  - Show: servers, chunk distribution, failures

- **Kubernetes Deployment**
  - Auto-scale chunk servers

- **S3 Gateway**
  - Expose API like: PUT /file, GET /file

## Suggested Timeline

| Week | Goal |
|------|------|
| 1 | Basic master + chunk server |
| 2 | Chunk replication |
| 3 | gRPC communication |
| 4 | Fault tolerance |
| 5 | Persistence |
| 6 | Snapshots |
| 7 | Deployment + monitoring |