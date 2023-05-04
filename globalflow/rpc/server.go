package rpc

import (
	"context"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

// Server describes an RPC server.
type Server struct {
	// Address is the address to listen on.
	Address string

	// RPC is the RPC interface.
	rpc interface{}
}

// NewServer creates a new RPC server.
func NewServer(addr string, rpc interface{}) *Server {
	return &Server{
		Address: addr,
		rpc:     rpc,
	}
}

// Run runs the RPC server.
func (server *Server) Run() error {
	l, err := net.Listen("tcp", server.Address)
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
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	//defer cancel()
	//
	//return s.Shutdown(ctx)

	return nil
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		logrus.WithError(err).Error("failed to accept websocket")

		return
	}

	go func(c *websocket.Conn) {
		server.readSocket(c)
	}(c)
}

func (server *Server) readSocket(c *websocket.Conn) {
	for {
		t, m, err := c.Read(context.Background())
		if err != nil {
			logrus.WithError(err).Warn("failed to read from websocket")

			break
		}

		if t == websocket.MessageText {
			req, err := Decode(m)
			if err != nil {
				logrus.WithError(err).Warn("failed to decode message")

				continue
			}

			switch req.(type) {
			case RequestMessage:

			}
		}
	}
}
