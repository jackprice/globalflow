package globalflow

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"nhooyr.io/websocket"
)

type MessageType string

const (
	MessageTypeCommand MessageType = "command"
)

type Message interface {
	MessageType() MessageType
	GetOriginator() string
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
}

func (CommandMessage) MessageType() MessageType {
	return MessageTypeCommand
}

func (message CommandMessage) GetOriginator() string {
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

func (server *Server) broadcast(message Message) error {
	encoded, err := encodeMessage(message)
	if err != nil {
		return err
	}

	next := server.NextLocalNode()
	if next != nil && next.NodeID() != message.GetOriginator() {
		logrus.Debug("broadcasting to local node")

		c, err := server.GetSocket(next)
		if err == nil {
			c.Write(context.Background(), websocket.MessageText, encoded)
		} else {
			logrus.Error(err)
		}
	}

	return nil
}
