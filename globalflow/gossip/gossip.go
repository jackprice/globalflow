package gossip

import (
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"globalflow/config"
	"time"
)

// Gossip is a gossip protocol.
type Gossip struct {
	configuration *config.Configuration
	m             *memberlist.Memberlist
	events        *EventDelegate
	messageCh     chan []byte
}

// NewGossip creates a new gossip protocol.
func NewGossip(cfg *config.Configuration) (*Gossip, error) {
	return &Gossip{
		configuration: cfg,
		events:        NewEventDelegate(),
		messageCh:     make(chan []byte, 128),
	}, nil
}

// Start starts the gossip protocol.
func (g *Gossip) Start() error {
	logrus.Debug("Starting gossip server")

	cfg := memberlist.DefaultLANConfig()

	cfg.Name = g.configuration.NodeID
	cfg.BindPort = g.configuration.NodePort
	//cfg.AdvertiseAddr = g.configuration.NodeAddress
	cfg.AdvertisePort = g.configuration.NodePort
	cfg.LogOutput = &LogrusLogger{}
	cfg.Delegate = &Delegate{
		Metadata: GossipMetadata{
			Region:   g.configuration.NodeRegion,
			Zone:     g.configuration.NodeZone,
			Hostname: g.configuration.NodeHostname,
		},
		MessageChan: g.messageCh,
	}
	cfg.Events = g.events

	m, err := memberlist.Create(cfg)
	if err != nil {
		return err
	}
	g.m = m

	if len(g.configuration.NodePeers) > 0 {
		go func() {
			for {
				_, err := m.Join(g.configuration.NodePeers)
				if err != nil {
					logrus.WithError(err).Error("Failed to join cluster")
				} else {
					break
				}

				time.Sleep(time.Second * 10)
			}
		}()
	}

	return nil
}

// Close closes the gossip protocol.
func (g *Gossip) Close() (err error) {
	if err := g.m.Leave(time.Second * 60); err != nil {
		logrus.WithError(err).Error("Failed to broadcast leave message")
	}

	_ = g.m.Shutdown()

	return
}

// Members returns the members in the cluster. This can include the local node, and suspect nodes.
func (g *Gossip) Members() []*memberlist.Node {
	return g.m.Members()
}

// MessageCh returns a channel that can be listened to to receive messages from the cluster.
func (g *Gossip) MessageCh() chan []byte {
	return g.messageCh
}

// SendReliable reliably sends a message to a node.
func (g *Gossip) SendReliable(to *memberlist.Node, msg []byte) (err error) {
	// Retry sending the message 3 times.
	for i := 0; i < 3; i++ {
		err := g.m.SendReliable(to, msg)
		if err == nil {
			return nil
		}
	}

	return err
}
