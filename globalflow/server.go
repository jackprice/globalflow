package globalflow

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/redcon"
	"globalflow/globalflow/db"
	"globalflow/globalflow/gossip"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"path"
	"sync"
	"time"
)

type Server struct {
	// container is the IoC container.
	container *Container

	// sockets contains websocket connections for all connected nodes.
	// It's a map of node name to websocket connection.
	sockets map[string]*websocket.Conn

	// socketMutex is a mutex for socketMutexes.
	// It must be held when writing to socketMutexes.
	socketMutex sync.Mutex

	// socketMutexes contains mutexes for connecting to nodes.
	socketMutexes map[string]*sync.Mutex

	// channels contains channels for communicating with other nodes.
	channels Channels

	// db is the database.
	db *db.Database

	// clock contains a Lamport clock.
	clock *LamportCLock

	// gossip contains the Gossip protocol implementation
	gossip *gossip.Gossip

	// shutdownCh is a channel for shutting down the server.
	shutdownCh chan struct{}

	// httpServer is the http server.
	httpServer *http.Server
}

// Channels contains channels for communicating with other nodes.
type Channels struct{}

const NodeNameHTTPHeader = "X-Node-Name"

// NewServer creates a new server.
func NewServer(container *Container) *Server {
	return &Server{
		container:     container,
		sockets:       make(map[string]*websocket.Conn),
		socketMutexes: make(map[string]*sync.Mutex),
		channels:      Channels{},
		clock:         NewClock(),
	}
}

// Run runs the server until terminated.
func (server *Server) Run(ctx context.Context) error {
	dbPath := path.Join(server.container.Configuration.DatabasePath, fmt.Sprintf("%s.db", server.container.Configuration.NodeID))

	db, err := db.NewDatabase(dbPath)
	if err != nil {
		return err
	}

	server.db = db

	err = server.StartGossip()
	if err != nil {
		return err
	}

	go func() {
		err := redcon.ListenAndServe(
			fmt.Sprintf(":%d", server.container.Configuration.RedisPort),
			server.Redis,
			func(conn redcon.Conn) bool {
				logrus.Debugf("Accepted connection from %s", conn.RemoteAddr())

				return true
			},
			func(conn redcon.Conn, err error) {},
		)
		if err != nil {
			logrus.WithError(err).Error("failed to start redis server")
		}
	}()

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", server.container.Configuration.NodePort))
	if err != nil {
		return err
	}

	logrus.Infof("Listening on %s", l.Addr().String())

	s := &http.Server{
		Handler:      server,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	server.httpServer = s

	return s.Serve(l)
}

func (server *Server) Close() error {
	logrus.Info("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if server.httpServer != nil {
		err := server.httpServer.Shutdown(ctx)
		if err != nil {
			logrus.Warn("failed to close http server cleanly")
		}
	}

	if server.gossip != nil {
		err := server.gossip.Close()
		if err != nil {
			logrus.Warn("failed to close gossip cleanly")
		}
	}

	if server.db != nil {
		err := server.db.Close()
		if err != nil {
			logrus.Warn("failed to close database cleanly")
		}
	}

	return nil
}

// ServeHTTP serves HTTP requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if server.gossip == nil {
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"default", "stream", "packet"}})
	if err != nil {
		logrus.WithError(err).Error("failed to accept websocket")

		return
	}

	logrus.WithField("addr", r.RemoteAddr).WithField("protocol", c.Subprotocol()).Debug("Accepted websocket connection")

	if c.Subprotocol() == "stream" || c.Subprotocol() == "packet" {
		server.gossip.HandleConnection(c, r)
	} else {
		server.readSocket(c)
	}
}

func (server *Server) readSocket(c *websocket.Conn) {
	for {
		t, msg, err := c.Read(context.Background())
		if err != nil {
			logrus.WithError(err).Warn("failed to read from websocket")

			break
		}

		if t == websocket.MessageText {
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

func (server *Server) handleCommand(cmd *CommandMessage) {
	server.clock.Set(cmd.Time)

	server.processCommand(cmd)

	cmd.DecrementTTL()

	if cmd.GetTTL() > 0 {
		err := server.broadcast(cmd)
		if err != nil {
			logrus.WithError(err).Warn("failed to broadcast command")
		}
	}
}

func (server *Server) processCommand(cmd *CommandMessage) {
	// TODO: Write to log

	switch cmd.Command {
	case "set":
		err := server.db.Set(time.Now(), cmd.Arguments[0], cmd.Arguments[1], 0)

		if err != nil {
			logrus.WithError(err).Warn("failed to set key")
		}

	case "del":
		err := server.db.Delete(cmd.Arguments[0])

		if err != nil {
			logrus.WithError(err).Warn("failed to delete key")
		}

	default:
		logrus.Warnf("Unknown command: %s", cmd.Command)
	}
}

// GetSocket gets a socket for a node.
// TODO: This should do some kind of connection pooling.
func (server *Server) GetSocket(node *Node) (*websocket.Conn, error) {
	server.socketMutex.Lock()

	c, ok := server.sockets[node.NodeID()]
	if ok {
		server.socketMutex.Unlock()
		return c, nil
	}

	m, ok := server.socketMutexes[node.NodeID()]
	if !ok {
		m = &sync.Mutex{}
		server.socketMutexes[node.NodeID()] = m
	}
	server.socketMutex.Unlock()

	m.Lock()
	defer m.Unlock()

	logrus.WithField("addr", node.Address()).Debug("Dialing websocket")

	c, _, err := websocket.Dial(context.Background(), fmt.Sprintf("ws://%s", node.Address()), &websocket.DialOptions{Subprotocols: []string{"default"}})
	if err != nil {
		return nil, err
	}

	server.sockets[node.NodeID()] = c

	return c, nil
}
