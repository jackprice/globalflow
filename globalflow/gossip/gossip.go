package gossip

import (
	"fmt"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"globalflow/config"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

// Gossip is a gossip protocol.
type Gossip struct {
	transport     *Transport
	configuration *config.Configuration
	m             *memberlist.Memberlist
	events        *EventDelegate
}

// NewGossip creates a new gossip protocol.
func NewGossip(cfg *config.Configuration) (*Gossip, error) {
	transport := NewTransport(cfg.NodeID, fmt.Sprintf("%s:%d", cfg.NodeAddress, cfg.NodePort))

	return &Gossip{
		transport:     transport,
		configuration: cfg,
		events:        NewEventDelegate(),
	}, nil
}

// Start starts the gossip protocol.
func (g *Gossip) Start() error {
	logrus.Debug("Starting gossip server")

	cfg := memberlist.DefaultWANConfig()

	cfg.Name = g.configuration.NodeID
	cfg.BindPort = g.configuration.NodePort
	cfg.AdvertiseAddr = g.configuration.NodeAddress
	cfg.AdvertisePort = g.configuration.NodePort
	cfg.Transport = g.transport
	cfg.LogOutput = &LogrusLogger{}
	cfg.Delegate = &Delegate{
		Metadata: GossipMetadata{
			Region:   g.configuration.NodeRegion,
			Zone:     g.configuration.NodeZone,
			Hostname: g.configuration.NodeHostname,
		},
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

// HandleConnection handles a websocket connection.
// This should be called by the server after it has determined a websocket connection is destined for the gossip protocol.
func (g *Gossip) HandleConnection(conn *websocket.Conn, r *http.Request) {
	logrus.Tracef("Handling connection")

	g.transport.handleConnection(conn, r)
}

// Members returns the members in the cluster. This can include the local node, and suspect nodes.
func (g *Gossip) Members() []*memberlist.Node {
	return g.m.Members()
}
