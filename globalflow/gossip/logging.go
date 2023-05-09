package gossip

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

type LogrusLogger struct {
}

func (l LogrusLogger) Write(p []byte) (n int, err error) {
	parsed, err := ParseLogLine(string(p))
	if err != nil {
		return 0, err
	}

	// Convert the log level
	ll, err := logrus.ParseLevel(parsed.Level)
	if err != nil {
		return 0, err
	}

	logrus.StandardLogger().Log(ll, parsed.Message)

	return len(p), nil
}

type ParsedLog struct {
	Date      string
	Time      string
	Level     string
	Component string
	Message   string
}

// ParseLogLine parses a log line into a ParsedLog struct.
func ParseLogLine(line string) (*ParsedLog, error) {
	var components []string

	for {
		i := strings.IndexRune(line, ' ')

		substr := line[:i]
		line = line[i+1:]

		components = append(components, substr)

		if len(line) == 0 {
			break
		}

		if len(components) == 4 {
			components = append(components, line)
			break
		}
	}

	if len(components) != 5 {
		return nil, fmt.Errorf("invalid log line")
	}

	return &ParsedLog{
		Date:      components[0],
		Time:      components[1],
		Level:     components[2][1 : len(components[2])-1],
		Component: components[3][:len(components[3])-1],
		Message:   components[4],
	}, nil
}

var _ io.Writer = &LogrusLogger{}
