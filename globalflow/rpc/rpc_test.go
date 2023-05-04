package rpc

import (
	"fmt"
	"math/rand"
	"nhooyr.io/websocket"
	"testing"
	"time"
)

type TestRPC struct {
}

func TestRemoteHost(t *testing.T) {
	port := rand.Intn(32768) + 1024

	rpc := TestRPC{}

	server := NewServer(fmt.Sprintf(":%d", port), rpc)

	if err := server.Run(); err != nil {
		t.Fatal(err)
	}

	client := NewRemoteHost(fmt.Sprintf("ws://localhost:%d", port), websocket.DialOptions{})

	// Sleep for 1 second to allow the server to start.
	time.Sleep(time.Second)

	err := client.Call("TestRPC.Test", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 10)
}
