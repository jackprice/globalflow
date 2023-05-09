package globalflow

import (
	"github.com/spaolacci/murmur3"
	"sort"
)

// RingIndex returns the ring index for a given node ID.
func RingIndex(nodeID string) uint32 {
	return murmur3.Sum32([]byte(nodeID))
}

// RingIndex returns the ring index for this node.
func (server *Server) RingIndex() uint32 {
	return RingIndex(server.container.Configuration.NodeID)
}

// RingIndex returns the ring index for this node.
func (n *Node) RingIndex() uint32 {
	return RingIndex(n.NodeID())
}

type ByRingIndex []*Node

func (a ByRingIndex) Len() int           { return len(a) }
func (a ByRingIndex) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRingIndex) Less(i, j int) bool { return a[i].RingIndex() < a[j].RingIndex() }

func (server *Server) LocalNodes() []*Node {
	nodes := make([]*Node, 0)

	for _, node := range server.Nodes() {
		if node.NodeID() == server.container.Configuration.NodeID {
			continue
		}

		metadata := node.Metadata()
		if metadata == nil {
			continue
		}

		if metadata.Region == server.container.Configuration.NodeRegion {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func (server *Server) RemoteNodes() []*Node {
	nodes := make([]*Node, 0)

	for _, node := range server.Nodes() {
		if node.NodeID() == server.container.Configuration.NodeID {
			continue
		}

		metadata := node.Metadata()
		if metadata == nil {
			continue
		}

		if metadata.Region != server.container.Configuration.NodeRegion {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// NextLocalNode returns the next node in the ring in the same datacenter.
func (server *Server) NextLocalNode() *Node {
	nodes := server.LocalNodes()

	sort.Sort(ByRingIndex(nodes))

	for _, node := range server.Nodes() {
		if node.RingIndex() > server.RingIndex() {
			return node
		}
	}

	if len(nodes) > 0 {
		return nodes[0]
	}

	return nil
}

// NextRemoteNode returns the next node in the ring in a different datacenter.
func (server *Server) NextRemoteNode() *Node {
	nodes := server.RemoteNodes()

	sort.Sort(ByRingIndex(nodes))

	for _, node := range server.Nodes() {
		if node.RingIndex() > server.RingIndex() {
			return node
		}
	}

	if len(nodes) > 0 {
		return nodes[0]
	}

	return nil
}
