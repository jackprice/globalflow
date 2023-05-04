package rpc

import (
	"encoding/json"
	"fmt"
)

const (
	MessageTypeRequest MessageType = "request"
	MessageTypeReply   MessageType = "reply"
)

type MessageType string

type Message interface {
	MessageType() MessageType
}

type RequestMessage struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params string `json:"params"`
}

type ReplyMessage struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}

type payload struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (RequestMessage) MessageType() MessageType {
	return MessageTypeRequest
}

func (ReplyMessage) MessageType() MessageType {
	return MessageTypeReply
}

func Encode(message Message) ([]byte, error) {
	p, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return json.Marshal(payload{
		Type:    message.MessageType(),
		Payload: p,
	})
}

type internalMessage struct {
	Type MessageType `json:"type"`
}

func Decode(data []byte) (Message, error) {
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	switch p.Type {
	case MessageTypeRequest:
		request := RequestMessage{}
		if err := json.Unmarshal(p.Payload, &request); err != nil {
			return nil, err
		}

		return request, nil
	}

	return nil, fmt.Errorf("unknown p type: %s", p.Type)
}
