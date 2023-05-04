package globalflow

import "encoding/json"

type MessageType string

const (
	MessageTypeHeartbeat   MessageType = "heartbeat"
	MessageTypeRequestVote MessageType = "request_vote"
	MessageTypeVodeGranted MessageType = "vote_granted"
)

type Message interface {
	MessageType() MessageType
}

type internalMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type HeartbeatMessage struct {
	NodeID string `json:"node_id"`
	Term   int    `json:"term"`
}

func (HeartbeatMessage) MessageType() MessageType {
	return MessageTypeHeartbeat
}

type RequestVoteMessage struct {
	NodeID string `json:"node_id"`
	Term   int    `json:"term"`
}

func (RequestVoteMessage) MessageType() MessageType {
	return MessageTypeRequestVote
}

type VoteGrantedMessage struct {
	NodeID string `json:"node_id"`
	Term   int    `json:"term"`
}

func (VoteGrantedMessage) MessageType() MessageType {
	return MessageTypeVodeGranted
}

// {"type": "vote_granted", "payload": {"node_id": "Bar"}}

// decodeMessage decodes a message from a byte slice.
func decodeMessage(data []byte) (interface{}, error) {
	var message internalMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	switch message.Type {
	case MessageTypeHeartbeat:
		return HeartbeatMessage{}, nil
	case MessageTypeRequestVote:
		var requestVote RequestVoteMessage
		if err := json.Unmarshal(message.Payload, &requestVote); err != nil {
			return nil, err
		}

		return requestVote, nil
	case MessageTypeVodeGranted:
		var voteGranted VoteGrantedMessage
		if err := json.Unmarshal(message.Payload, &voteGranted); err != nil {
			return nil, err
		}

		return voteGranted, nil
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
