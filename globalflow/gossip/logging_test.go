package gossip

import "testing"

func TestParseLogLine(t *testing.T) {
	line := "2023/05/05 16:03:42 [WARN] memberlist: Refuting"

	parsed, err := ParseLogLine(line)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Date != "2023/05/05" {
		t.Error("Invalid date")
	}

	if parsed.Time != "16:03:42" {
		t.Error("Invalid time")
	}

	if parsed.Level != "WARN" {
		t.Error("Invalid level")
	}

	if parsed.Component != "memberlist" {
		t.Error("Invalid component")
	}

	if parsed.Message != "Refuting" {
		t.Error("Invalid message")
	}
}
