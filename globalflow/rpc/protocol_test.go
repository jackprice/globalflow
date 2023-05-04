package rpc

import (
	"testing"
)

func TestEncode(t *testing.T) {
	request := RequestMessage{
		ID:     "",
		Method: "",
	}

	encoded, err := Encode(request)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.MessageType() != MessageTypeRequest {
		t.Fatalf("expected MessageTypeRequest, got %s", decoded.MessageType())
	}
}
