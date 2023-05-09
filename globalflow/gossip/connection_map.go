package gossip

import (
	"context"
	"fmt"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

// ConnectionMap is a mutex-backed map of connections.
type ConnectionMap struct {
	// connections contains connections to other nodes.
	connections map[string]*websocket.Conn

	// mu is a mutex for connections.
	// It must be held when writing to connections.
	mu sync.Mutex

	// Subprotocol is the Websocket subprotocol to connect to.
	Subprotocol string

	// HTTPHeader contains HTTP headers to send when connecting
	HTTPHeader http.Header
}

// NewConnectionMap creates a new connection map.
func NewConnectionMap(path string) *ConnectionMap {
	return &ConnectionMap{
		connections: make(map[string]*websocket.Conn),
		Subprotocol: path,
		HTTPHeader:  make(http.Header),
	}
}

// getConnection gets a connection to a node.
func (c *ConnectionMap) getConnection(addr string) (*websocket.Conn, error) {
	// TODO: This locking is not ideal - it locks everything for the duration of the dial.
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.connections[addr]
	if ok {
		return conn, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/%s", addr, c.Subprotocol), &websocket.DialOptions{HTTPHeader: c.HTTPHeader, Subprotocols: []string{c.Subprotocol}})
	if err != nil {
		return nil, err
	}

	c.connections[addr] = conn
	return conn, nil
}

func (c *ConnectionMap) removeConnection(addr string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.connections, addr)
}
