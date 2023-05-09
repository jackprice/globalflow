package gossip

import (
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"sync"
)

type EventDelegate struct {
	Members map[string]*memberlist.Node

	mu sync.Mutex
}

func (e *EventDelegate) NotifyJoin(node *memberlist.Node) {
	logrus.WithField("node", node.Name).Debug("Node joined")

	e.mu.Lock()
	defer e.mu.Unlock()

	e.Members[node.Name] = node
}

func (e *EventDelegate) NotifyLeave(node *memberlist.Node) {
	logrus.WithField("node", node.Name).Debug("Node left")

	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.Members, node.Name)
}

func (e *EventDelegate) NotifyUpdate(node *memberlist.Node) {
	logrus.WithField("node", node.Name).Debug("Node updated")

	e.mu.Lock()
	defer e.mu.Unlock()

	e.Members[node.Name] = node
}

// NewEventDelegate creates a new event delegate.
func NewEventDelegate() *EventDelegate {
	return &EventDelegate{
		Members: make(map[string]*memberlist.Node),
	}
}

var _ memberlist.EventDelegate = &EventDelegate{}
