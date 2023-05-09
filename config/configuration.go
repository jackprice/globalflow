package config

import "os"

// Configuration is a struct that contains the global configuration.
type Configuration struct {
	// DatabasePath is the path to the database file.
	DatabasePath string

	// NodeID is the ID of the node.
	NodeID string

	// NodeAddress is the address of the node.
	NodeAddress string

	// NodePort is the port of the node.
	NodePort int

	// NodePeers is a list of peers.
	NodePeers []string

	// NodeRegion is the region of the node.
	NodeRegion string

	// NodeZone is the zone of the node.
	NodeZone string

	// NodeHostname is the hostname of the node.
	NodeHostname string

	// RedisPort is the port to run the Redis server on.
	RedisPort int
}

// NewConfiguration creates a new configuration with default values.
func NewConfiguration() *Configuration {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	return &Configuration{
		DatabasePath: ".",
		NodePeers:    []string{},
		NodeAddress:  "localhost",
		NodeRegion:   "local",
		NodeZone:     "local",
		NodeHostname: hostname,
		RedisPort:    63790,
	}
}
