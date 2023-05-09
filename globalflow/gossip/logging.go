package gossip

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
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
	var date string
	var time string
	var level string
	var component string
	var message string

	_, err := fmt.Sscanf(line, "%s %s %s %s %s", &date, &time, &level, &component, &message)
	if err != nil {
		return nil, err
	}

	// Trim the leading and trailing [] from the level
	level = level[1 : len(level)-1]

	// Trim the trailing : from the component
	component = component[:len(component)-1]

	return &ParsedLog{
		Date:      date,
		Time:      time,
		Level:     level,
		Component: component,
		Message:   message,
	}, nil
}

var _ io.Writer = &LogrusLogger{}
