package gossip

// GossipMetadata contains metadata for nodes.
type GossipMetadata struct {
	Region   string `json:"region"`
	Zone     string `json:"zone"`
	Hostname string `json:"hostname"`
}
