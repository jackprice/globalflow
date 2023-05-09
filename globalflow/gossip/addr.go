package gossip

import (
	"fmt"
	"net"
)

type WSAddr struct {
	Address string
	NodeID  string
}

func (w *WSAddr) Network() string {
	return "ws"
}

func (w *WSAddr) String() string {
	return fmt.Sprintf("%s (%s)", w.NodeID, w.Address)
}

var _ net.Addr = &WSAddr{}
