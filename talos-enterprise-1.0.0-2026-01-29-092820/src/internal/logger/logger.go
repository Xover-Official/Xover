package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Agent string

const (
	Architect  Agent = "Architect"
	Auditor    Agent = "Auditor"
	Builder    Agent = "Builder"
	Strategist Agent = "Strategist"
	Sentinel   Agent = "Sentinel"
)

type LogEntry struct {
	Timestamp time.Time     `json:"timestamp"`
	Agent     Agent         `json:"agent"`
	Action    string        `json:"action"`
	Status    string        `json:"status"`
	Metadata  string        `json:"metadata,omitempty"`
	Latency   time.Duration `json:"latency,omitempty"`
	Tokens    int           `json:"tokens,omitempty"`
}

func LogAction(agent Agent, action, status, metadata string) error {
	return LogFullAction(agent, action, status, metadata, 0, 0)
}

func LogFullAction(agent Agent, action, status, metadata string, latency time.Duration, tokens int) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Agent:     agent,
		Action:    action,
		Status:    status,
		Metadata:  metadata,
		Latency:   latency,
		Tokens:    tokens,
	}

	f, err := os.OpenFile("SESSION_LOG.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(entry); err != nil {
		return err
	}

	fmt.Printf("[%s] %s: %s (%s)\n", entry.Agent, entry.Action, entry.Status, entry.Metadata)
	return nil
}
