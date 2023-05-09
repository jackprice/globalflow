package globalflow

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/memberlist"
)

// Node is a wrapper around a remote node.
type Node struct {
	// node is the memberlist node.
	node *memberlist.Node

	// metadata is the gossip metadata if it has been decoded.
	metadata *GossipMetadata
}

// NewNode creates a new node.
func NewNode(node *memberlist.Node) *Node {
	return &Node{
		node: node,
	}
}

// Metadata returns the gossip metadata for the node.
func (n *Node) Metadata() *GossipMetadata {
	if n.metadata != nil {
		return n.metadata
	}

	if n.node.Meta == nil {
		return nil
	}

	metadata := &GossipMetadata{}

	err := json.Unmarshal(n.node.Meta, metadata)
	if err != nil {
		return nil
	}

	n.metadata = metadata

	return metadata
}

// NodeID returns the node ID.
func (n *Node) NodeID() string {
	return n.node.Name
}

func (n *Node) Address() string {
	return fmt.Sprintf("%s:%d", n.node.Addr.String(), n.node.Port)
}

// Nodes returns the nodes in the cluster.
func (server *Server) Nodes() []*Node {
	nodes := make([]*Node, 0)

	for _, node := range server.gossip.Members() {
		nodes = append(nodes, NewNode(node))
	}

	return nodes
}
