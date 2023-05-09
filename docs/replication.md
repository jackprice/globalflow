# Replication

Globalflow automatically replicates data with a probabilistic consistency model.

## Mechanics

Globalflow uses a gossip protocol to discover the state of the network. Each node maintains a list of all other nodes
in the network and their topology.

Nodes are _ordered_ in a ring. Each node is aware of its successor in the ring, which may change depending on which 
nodes are available. The ring is used to coordinate the replication of data.

## Replication

When a node receives a write request, it writes the data to its local storage and then forwards the request to
1. Its first successor in the current availability zone
2. Its first successor in the next availability zone

![Ring architecture](./ring-architecture.png)