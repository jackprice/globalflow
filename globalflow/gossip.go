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

	go server.StreamMessages()

	return nil
}

// StreamMessages starts handling messages from the gossip server.
func (server *Server) StreamMessages() {
	for {
		select {
		case msg := <-server.gossip.MessageCh():
			decoded, err := decodeMessage(msg)
			if err != nil {
				logrus.WithError(err).Warn("failed to decode message")
			}

			logrus.Debugf("Received message: %T", decoded)

			switch v := decoded.(type) {
			case CommandMessage:
				server.handleCommand(&v)

			default:
				logrus.Warnf("Unknown message type: %T", v)
			}
		}
	}
}
