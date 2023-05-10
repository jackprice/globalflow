package globalflow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
	"strings"
)

type MessageType string

const (
	MessageTypeCommand MessageType = "command"
)

type Message interface {
	MessageType() MessageType
	GetOriginator() string
	GetTTL() int
	DecrementTTL()
}

type internalMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type CommandMessage struct {
	Time       Time     `json:"clock"`
	Command    string   `json:"command"`
	Arguments  []string `json:"arguments"`
	Originator string   `json:"originator"`
	TTL        int      `json:"ttl"`
}

func (CommandMessage) MessageType() MessageType {
	return MessageTypeCommand
}

func (message *CommandMessage) GetTTL() int {
	return message.TTL
}

func (message *CommandMessage) DecrementTTL() {
	message.TTL--
}

func (message *CommandMessage) GetOriginator() string {
	return message.Originator
}

// decodeMessage decodes a message from a byte slice.
func decodeMessage(data []byte) (interface{}, error) {
	var message internalMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	switch message.Type {
	case MessageTypeCommand:
		var command CommandMessage
		if err := json.Unmarshal(message.Payload, &command); err != nil {
			return nil, err
		}

		return command, nil
	}

	return nil, nil
}

// encodeMessage encodes a message to a byte slice.
func encodeMessage(message Message) ([]byte, error) {
	var internal internalMessage
	internal.Type = message.MessageType()

	payload, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	internal.Payload = payload

	return json.Marshal(internal)
}

func (server *Server) NewCommandMessage(command string, arguments []string) *CommandMessage {
	return &CommandMessage{
		Time:       server.clock.Get(),
		Command:    strings.ToLower(command),
		Arguments:  arguments,
		Originator: server.container.Configuration.NodeID,
		TTL:        len(server.gossip.Members()) + 1,
	}
}

// broadcast broadcasts a message to other nodes.
// Returns an error if no nodes are available.
func (server *Server) broadcast(message Message) error {
	logrus.WithField("ttl", message.GetTTL()).Debugf("broadcasting message: %s", message)

	encoded, err := encodeMessage(message)
	if err != nil {
		return err
	}

	count := 0

	next := server.NextLocalNode()
	if next != nil {
		logrus.Debug("broadcasting to local node")

		err := server.gossip.SendReliable(next.node, encoded)

		if err == nil {
			count++
		} else {
			logrus.Error(err)
		}
	}

	nextRemote := server.NextRemoteNode()
	if nextRemote != nil {
		logrus.Debug("broadcasting to remote node")

		c, err := server.GetSocket(nextRemote)
		if err == nil {
			err := c.Write(context.Background(), websocket.MessageText, encoded)
			if err == nil {
				count++
			}
		} else {
			logrus.Error(err)
		}
	}

	if count == 0 {
		return fmt.Errorf("no nodes available")
	}

	return nil
}
