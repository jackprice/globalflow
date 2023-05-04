package rpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

type RemoteHost struct {
	// Address contains the address of the remote host.
	Address string

	// DialOptions contains the dial options for the remote host.
	DialOptions websocket.DialOptions

	// connections contains all websocket connections to the remote host.
	// Some of these may be closed or in a bad state.
	connections []*websocket.Conn

	// connectionMutex is a mutex for connections.
	// It must be held when writing to connections.
	connectionMutex sync.Mutex

	// roundRobinIndex is an incrementing index used to select a connection in a round-robin fashion.
	roundRobinIndex int
}

// NewRemoteHost creates a new remote host.
func NewRemoteHost(address string, dialOptions websocket.DialOptions) *RemoteHost {
	return &RemoteHost{
		Address:     address,
		DialOptions: dialOptions,
	}
}

// addConnection adds a new connection to the remote host.
func (h *RemoteHost) addConnection(conn *websocket.Conn) {
	h.connectionMutex.Lock()
	defer h.connectionMutex.Unlock()

	h.connections = append(h.connections, conn)
}

// removeConnection removes a connection from the remote host.
func (h *RemoteHost) removeConnection(conn *websocket.Conn) {
	h.connectionMutex.Lock()
	defer h.connectionMutex.Unlock()

	for i, c := range h.connections {
		if c == conn {
			h.connections = append(h.connections[:i], h.connections[i+1:]...)
			return
		}
	}
}

// getConnection gets an available connection to the remote host. If none exist, a new one is created.
func (h *RemoteHost) getConnection() (*websocket.Conn, error) {
	// TODO: Configure max connections
	if len(h.connections) < 4 {
		// TODO: Record HTTP response somewhere
		c, _, err := websocket.Dial(context.Background(), h.Address, &h.DialOptions)
		if err != nil {
			return nil, err
		}

		h.addConnection(c)

		return c, nil
	}

	h.roundRobinIndex++

	return h.connections[h.roundRobinIndex%len(h.connections)], nil
}

func (h *RemoteHost) Call(method string, in interface{}, out interface{}) error {
	c, err := h.getConnection()
	if err != nil {
		return err
	}

	id := "foo"

	msg := RequestMessage{
		Method: method,
		ID:     id,
	}

	encoded, err := Encode(msg)
	if err != nil {
		return err
	}

	err = c.Write(context.Background(), websocket.MessageText, encoded)
	if err != nil {
		return err
	}

	deadline, _ := context.WithTimeout(context.Background(), 10*time.Second)
	t, m, err := c.Read(deadline)
	if err != nil {
		return err
	}

	if t != websocket.MessageText {
		return fmt.Errorf("expected text message, got %s", t)
	}

	p, err := Decode(m)
	if err != nil {
		return err
	}

	pm, ok := p.(ReplyMessage)
	if !ok {
		return fmt.Errorf("expected reply message, got %T", p)
	}

	if pm.ID != id {
		return fmt.Errorf("expected reply message ID %s, got %s", id, pm.ID)
	}

	encoder := gob.NewDecoder(bytes.NewBufferString(pm.Result))

	err = encoder.Decode(out)
	if err != nil {
		return err
	}

	return nil
}
