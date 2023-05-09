package globalflow

import (
	"github.com/sirupsen/logrus"
	"globalflow/globalflow/gossip"
)

// GossipMetadata contains metadata for nodes.
type GossipMetadata struct {
	Region   string `json:"region"`
	Zone     string `json:"zone"`
	Hostname string `json:"hostname"`
}

// StartGossip starts the gossip server
func (server *Server) StartGossip() error {

	logrus.Debug("Starting gossip server")

	g, err := gossip.NewGossip(server.container.Configuration)
	if err != nil {
		return err
	}

	err = g.Start()
	if err != nil {
		return err
	}

	server.gossip = g

	return nil
}
