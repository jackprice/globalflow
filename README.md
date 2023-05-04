# GlobalFlow

A _very_ experimental Redis-compatible globally distributed KV store written in Go.

It's intended for low latency reads and high availability with eventual consistency.

## Objectives

- To be simple to reason about
- To be simple to deploy and use - the system will handle all the complexity
- Low latency for reads
- High availability
- Be Redis-compatible
