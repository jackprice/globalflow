package gossip

import (
	"context"
	"fmt"
	"net"
	"nhooyr.io/websocket"
	"time"
)

type WSConn struct {
	Remote *WSAddr
	Local  *WSAddr
	Conn   *websocket.Conn
	ch     chan []byte

	readDeadline  time.Time
	writeDeadline time.Time
}

func (w *WSConn) Read(b []byte) (n int, err error) {
	ctx, cancel := context.WithDeadline(context.Background(), w.readDeadline)
	defer cancel()

	select {
	case in := <-w.ch:
		return copy(b, in), nil
	case <-ctx.Done():
		return 0, fmt.Errorf("failed to read: %w", ctx.Err())
	}
}

func (w *WSConn) Write(b []byte) (n int, err error) {
	ctx, cancel := context.WithDeadline(context.Background(), w.writeDeadline)
	defer cancel()

	err = w.Conn.Write(ctx, websocket.MessageBinary, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (w *WSConn) Close() error {
	return w.Conn.Close(websocket.StatusNormalClosure, "closing")
}

func (w *WSConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (w *WSConn) RemoteAddr() net.Addr {
	return w.Remote
}

func (w *WSConn) SetDeadline(t time.Time) error {
	w.readDeadline = t
	w.writeDeadline = t

	return nil
}

func (w *WSConn) SetReadDeadline(t time.Time) error {
	w.readDeadline = t

	return nil
}

func (w *WSConn) SetWriteDeadline(t time.Time) error {
	w.writeDeadline = t

	return nil
}

var _ net.Conn = &WSConn{}
