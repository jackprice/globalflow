package globalflow

import "testing"

func TestCommandMessage_DecrementTTL(t *testing.T) {
	msg := CommandMessage{
		TTL: 10,
	}

	msg.DecrementTTL()

	if msg.TTL != 9 {
		t.Errorf("Expected TTL to be 9, got %d", msg.TTL)
	}
}
