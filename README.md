# 🗄️ Distributed Godis

**Distributed Godis** is a distributed key-value store built in Go, aiming to explore the complexities and nuances of distributed systems. The project involves spinning up a cluster of nodes (for example, through bootstrapping) , each capable of handling requests and maintaining data consistency across the system.

---

## 🚀 Project Goals

- Develop a distributed key-value store 
- Implement core distributed systems concepts such as:
  - Cluster formation and node communication
  - Data replication and consistency mechanisms
  - Fault tolerance and recovery strategies
- Provide a simple API for interacting with the key-value store

---

## 🧱 Architecture Overview

The system is designed around a cluster of nodes, each running an instance of the Godis server. Nodes communicate with each other to replicate data and ensure consistency using Raft. The architecture includes:

- **API Layer**: Handles client requests for setting and retrieving key-value pairs. (In development)
- **Storage Engine**: Manages in-memory storage of key-value data.
- **Replication Mechanism**: Ensures data is replicated across nodes for fault tolerance.
- **Cluster Management**: Handles node discovery, health checks, and cluster state management.

---


