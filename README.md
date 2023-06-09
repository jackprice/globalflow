# GlobalFlow

A _very_ experimental Redis-compatible globally distributed KV store written in Go. 

It's intended for low latency reads and high availability with eventual consistency.

## Objectives

- To be simple to deploy and use - the system will hide all the complexity
- No cluster management - the system will automatically discover and manage the cluster
- To be simple to reason about
- Low latency for reads
- High availability
- Be Redis-compatible

## Assumptions

- Network communication is cheap and easy
- Clocks are synchronised (this is a pretty safe assumption)

## Communication

For ease of deployment across heterogeneous networks, GlobalFlow uses HTTP for inter-region communication. Low-latency 
intra-region communication is over TCP and UDP.

Globalflow is architected as a ring of rings. Each zone is a ring of nodes. Communication consistently flows in one direction.

![Ring architecture](./docs/ring-architecture.jpg)

## Reliability

GlobalFlows reliability model is probabilistic rather than deterministic. When you write a value to the store, it
is _probably_ persisted. When you read a value from the store, you will _probably_ get the latest value.

## Redis compatibility

The following Redis commands are supported:

- GET
- SET
- DEL
- LPUSH