Build a distributed system using 10 VMs from scratch in Golang - a log gripper (MP1), a failure detector(MP2), a distributed file system(MP3), and a distributed transaction system (MP4)

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

