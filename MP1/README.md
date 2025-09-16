# Distributed Log Querier - CS425 MP1

A distributed system for querying log files across multiple virtual machines using Go and RPC. This system allows you to execute grep commands across 10 VMs simultaneously and aggregate the results.

```
Client VM                    Server VMs (01-10)
┌─────────────┐             ┌─────────────┐
│ RPC Client  │────RPC─────▶│ RPC Server  │
│ Query       │   Calls     │ Grep Handler│
│ Aggregator  │◀───────────▶│ Log Files   │
└─────────────┘   Results   └─────────────┘
```
- **Communication**: RPC over HTTP on port 4425
- **Parallelism**: Simultaneous queries to all VMs
- **Fault Tolerance**: Graceful handling of VM failures


## Table of Contents
- [System Overview](#system-overview)
- [Project Structure](#project-structure)
- [Setup](#setup)
- [How to Run](#how-to-run)
- [Usage Examples](#usage-examples)


## System Overview

This distributed log querier consists of:
- **Client**: Sends grep queries to multiple VMs and aggregates results
- **Server**: RPC server on each VM that executes grep commands on local log files
- **Management Tools**: Scripts for VM startup/shutdown and repository synchronization

## Project Structure

```
cs-425-mp-1/
├── main/
│   └── main.go          # starts the client
├── client/
│   └── client.go        # RPC client implementation
├── server/
│   └── client.go        # RPC server implementation
├── startup/
│   └── startup.go       # for VM management utilities
├── tests/
│   └── unit_tests.go    # unit  tests for MP1
├── log/
│   └── vm.log           # Log files for testing
└── README.md
```


## Setup

### 1. Clone the Repository

```bash
# On each VM and your local machine
git clone https://github.com/mjwu106/Distributed-System/MP1.git
cd MP1
```

### 2. Configure SSH Keys

Ensure your SSH private key is configured in `startup.go`:
```go
privateKeyPath := "/path/to/your/.ssh/id_ed25519"  // Update this path
```

## How to Run

### Method 1: Using Scripts (Recommended)

#### Step 1: Start all VM servers
```bash
go run startup.go wake
```
This command will:
- SSH into all 10 VMs
- Kill any existing server processes
- Start the RPC server on each VM (port 4425)

#### Step 2: Run the client
```bash
go run main.go
```

#### Step 3: Stop all servers (when done)
```bash
go run startup.go kill
```

### Method 2: Manual

#### On each VM (01-10):
```bash
# SSH into each VM
ssh [netid]@fa25-cs425-b4[XX].cs.illinois.edu

# Navigate to project directory
cd ~/cs-425-mp-1

# Start the server
go run server.go
```

#### On client machine:
```bash
go run main.go
```

### Run Unit Tests
```bash
cd tests/
go run unit_tests.go
```

