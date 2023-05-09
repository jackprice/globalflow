package gossip

import (
	"context"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

type Transport struct {
	// streamConnections contains connections to other nodes.
	streamConnections *ConnectionMap

	// pack
	packetConnections *ConnectionMap

	// packetCh is a channel for receiving packets.
	packetCh chan *memberlist.Packet

	// streamCh is a channel for receiving streams.
	streamCh chan net.Conn

	nodeAddr string

	nodeID string
}

const NodeIDHttpHeader = "X-Node-Id"
const NodeAddrHttpHeader = "X-Node-Addr"

func (t *Transport) WriteToAddress(b []byte, addr memberlist.Address) (time.Time, error) {
	conn, err := t.packetConnections.getConnection(addr.Addr)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	err = conn.Write(context.Background(), websocket.MessageBinary, b)
	if websocket.CloseStatus(err) != -1 {
		logrus.Warn("Closing connection")

		t.packetConnections.removeConnection(addr.Addr)

		return time.Time{}, err
	}

	// TODO: Clean up connections that are closed here.

	return now, err
}

func (t *Transport) DialAddressTimeout(addr memberlist.Address, timeout time.Duration) (net.Conn, error) {
	return t.DialTimeout(addr.Addr, timeout)
}

func (t *Transport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	if ip == "" {
		ip = "127.0.0.1"
	}

	ips, err := net.LookupIP(ip)
	if err != nil {
		return nil, 0, err
	}
	if len(ips) > 0 {
		ip = ips[0].String()
	}

	return net.ParseIP(ip), port, nil
}

func (t *Transport) WriteTo(b []byte, addr string) (time.Time, error) {
	a := memberlist.Address{Addr: addr, Name: ""}
	return t.WriteToAddress(b, a)
}

func (t *Transport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

func (t *Transport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	// TODO: Respect timeout
	c, err := t.streamConnections.getConnection(addr)
	if err != nil {
		return nil, err
	}

	return websocket.NetConn(context.Background(), c, websocket.MessageBinary), nil

	//return &WSConn{
	//	Remote: &WSAddr{Address: addr},
	//	Local:  &WSAddr{Address: t.nodeAddr, NodeID: t.nodeID},
	//	Conn:   c,
	//}, nil
}

func (t *Transport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

func (t *Transport) Shutdown() error {
	// TODO: Implement me
	//for _, conn := range t.connections {
	//	conn.Close(websocket.StatusGoingAway, "shutting down")
	//}

	return nil
}

func (t *Transport) handleConnection(c *websocket.Conn, r *http.Request) {
	nodeID := r.Header.Get(NodeIDHttpHeader)
	if nodeID == "" {
		logrus.Warn("Received websocket connection without node ID")

		c.Close(websocket.StatusPolicyViolation, "missing node ID")

		return
	}

	nodeAddr := r.Header.Get(NodeAddrHttpHeader)
	if nodeAddr == "" {
		logrus.Warn("Received websocket connection without node address")

		c.Close(websocket.StatusPolicyViolation, "missing node address")

		return
	}

	if c.Subprotocol() == "stream" {
		//ch := make(chan []byte, 5)

		t.streamCh <- websocket.NetConn(context.Background(), c, websocket.MessageBinary)

		select {
		case <-r.Context().Done():
			logrus.Info("Context done")

			return
		}

		//t.streamCh <- &WSConn{
		//	Remote: &WSAddr{Address: nodeAddr, NodeID: nodeID},
		//	Local:  &WSAddr{Address: t.nodeAddr, NodeID: t.nodeID},
		//	Conn:   c,
		//	ch:     ch,
		//}
		//
		//for {
		//	tp, msg, err := c.Read(context.Background())
		//	if websocket.CloseStatus(err) != -1 {
		//		logrus.Debugf("Received close message from %s", nodeAddr)
		//	}
		//
		//	if err != nil {
		//		logrus.WithError(err).Error("failed to read stream websocket message")
		//
		//		return
		//	}
		//
		//	if tp != websocket.MessageBinary {
		//		logrus.WithError(err).Warn("received non-binary websocket message")
		//
		//		continue
		//	}
		//
		//	logrus.Debugf("Received stream message from %s", nodeAddr)
		//
		//	ch <- msg
		//}
	}

	for {
		tp, msg, err := c.Read(context.Background())
		now := time.Now()
		if websocket.CloseStatus(err) != -1 {
			logrus.Debugf("Received close message from %s", nodeAddr)
		}

		if err != nil {
			logrus.WithError(err).Error("failed to read packet websocket message")

			return
		}

		if tp != websocket.MessageBinary {
			logrus.WithError(err).Warn("received non-binary websocket message")

			continue
		}

		t.packetCh <- &memberlist.Packet{
			Buf: msg,
			From: &WSAddr{
				Address: nodeAddr,
				NodeID:  nodeID,
			},
			Timestamp: now,
		}
	}
}

func NewTransport(nodeID string, nodeAddr string) *Transport {
	t := &Transport{
		streamConnections: NewConnectionMap("stream"),
		packetConnections: NewConnectionMap("packet"),
		packetCh:          make(chan *memberlist.Packet),
		streamCh:          make(chan net.Conn),
		nodeAddr:          nodeAddr,
		nodeID:            nodeID,
	}

	t.streamConnections.HTTPHeader.Set(NodeIDHttpHeader, nodeID)
	t.streamConnections.HTTPHeader.Set(NodeAddrHttpHeader, nodeAddr)
	t.packetConnections.HTTPHeader.Set(NodeIDHttpHeader, nodeID)
	t.packetConnections.HTTPHeader.Set(NodeAddrHttpHeader, nodeAddr)

	return t
}

var _ memberlist.NodeAwareTransport = &Transport{}
