package globalflow

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"path"
	"sync"
	"time"
)

type Server struct {
	// Container is the IoC container.
	Container *Container

	// Sockets contains websocket connections for all connected nodes.
	// It's a map of node name to websocket connection.
	Sockets map[string]*websocket.Conn

	// socketMutex is a mutex for Sockets.
	// It must be held when writing to Sockets.
	socketMutex sync.Mutex

	// Channels contains channels for communicating with other nodes.
	Channels Channels

	// DB is the database.
	DB *bolt.DB
}

// Channels contains channels for communicating with other nodes.
type Channels struct {
	Heartbeat   chan HeartbeatMessage
	RequestVote chan RequestVoteMessage
	VoteGranted chan VoteGrantedMessage
}

// NewServer creates a new server.
func NewServer(container *Container) *Server {
	return &Server{
		Container: container,
		State:     NodeStateFollower,
		Auth:      &TokenAuth{token: "test"},
		Sockets:   make(map[string]*websocket.Conn),
		Channels: Channels{
			Heartbeat:   make(chan HeartbeatMessage),
			RequestVote: make(chan RequestVoteMessage),
			VoteGranted: make(chan VoteGrantedMessage),
		},
	}
}

// Run runs the server until terminated.
func (server *Server) Run(ctx context.Context) error {
	dbPath := path.Join(server.Container.Configuration.DatabasePath, fmt.Sprintf("%s.db", server.Container.Configuration.NodeID))

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}

	server.DB = db

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("DATA"))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	//go func() {
	//	err := redcon.ListenAndServe(
	//		":6381",
	//		server.Redis,
	//		func(conn redcon.Conn) bool {
	//			logrus.Debugf("Accepted connection from %s", conn.RemoteAddr())
	//
	//			return true
	//		},
	//		func(conn redcon.Conn, err error) {},
	//	)
	//	if err != nil {
	//		logrus.WithError(err).Error("failed to start redis server")
	//	}
	//}()

	l, err := net.Listen("tcp", ":9808")
	if err != nil {
		return err
	}

	logrus.Infof("Listening on %s", l.Addr().String())

	s := &http.Server{
		Handler:      server,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	errc := make(chan error, 1)

	go func() {
		errc <- s.Serve(l)
	}()

	server.StartRaft(ctx)

	server.ConnectToNode(ctx, "ws://localhost:8087")
	server.ConnectToNode(ctx, "ws://localhost:9809")

	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.Shutdown(ctx)
}

// ServeHTTP serves HTTP requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		logrus.WithError(err).Error("failed to accept websocket")

		return
	}

	nodeName := r.Header.Get(NodeNameHTTPHeader)
	if nodeName == "" {
		logrus.Error("missing node name")

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	server.setSocket(nodeName, c)

	go func(c *websocket.Conn) {
		server.readSocket(c)
	}(c)
}

func (server *Server) readSocket(c *websocket.Conn) {
	for {
		t, msg, err := c.Read(context.Background())
		if err != nil {
			logrus.WithError(err).Warn("failed to read from websocket")

			break
		}

		if t == websocket.MessageText {
			logrus.Infof("Received message: %s", msg)

			decoded, err := decodeMessage(msg)
			if err != nil {
				logrus.WithError(err).Warn("failed to decode message")
			}

			logrus.Debugf("Received message: %T", decoded)

			switch v := decoded.(type) {
			case HeartbeatMessage:
				server.Channels.Heartbeat <- v

			case VoteGrantedMessage:
				server.Channels.VoteGranted <- v

			case RequestVoteMessage:
				server.Channels.RequestVote <- v

			default:
				logrus.Warnf("Unknown message type: %T", v)
			}
		}
	}
}

func (server *Server) ConnectToNode(ctx context.Context, addr string) error {
	logrus.Infof("Connecting to %s", addr)

	c, r, err := websocket.Dial(ctx, addr, &websocket.DialOptions{
		HTTPHeader: http.Header{
			NodeNameHTTPHeader: []string{server.Container.Configuration.NodeID},
		},
	})
	if err != nil {
		logrus.WithError(err).Error("failed to dial websocket")

		return err
	}

	nodeID := r.Header.Get(NodeNameHTTPHeader)
	if nodeID == "" {
		//logrus.Error("missing node ID")
		//
		//return errors.New("missing node ID")
		nodeID = "server1"
	}

	server.setSocket(nodeID, c)

	go func(c *websocket.Conn) {
		server.readSocket(c)
	}(c)

	return nil
}

// setSocket sets a websocket connection by name.
func (server *Server) setSocket(name string, conn *websocket.Conn) {
	server.socketMutex.Lock()
	defer server.socketMutex.Unlock()

	// First check if there's an existing connection and close it
	if server.Sockets[name] != nil {
		err := server.Sockets[name].Close(websocket.StatusNormalClosure, "replaced by new connection")
		if err != nil {
			logrus.WithError(err).Error("failed to close existing websocket")
		}
	}

	server.Sockets[name] = conn
}

// getSocket gets a websocket connection by name. This may be nil.
func (server *Server) getSocket(name string) *websocket.Conn {
	server.socketMutex.Lock()
	defer server.socketMutex.Unlock()

	return server.Sockets[name]
}

func (server *Server) Broadcast(ctx context.Context, message Message) error {
	encoded, err := encodeMessage(message)
	if err != nil {
		logrus.WithError(err).Error("failed to encode message")

		return err
	}

	for name, conn := range server.Sockets {
		err := conn.Write(ctx, websocket.MessageText, encoded)
		if err != nil {
			logrus.WithError(err).Warningf("failed to write to %s", name)
		}
	}

	return nil
}

func (server *Server) Quorum() int {
	if len(server.Sockets) == 0 {
		return 1
	}

	if len(server.Sockets) == 1 {
		return 2
	}

	return len(server.Sockets)/2 + 1
}
