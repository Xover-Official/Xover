package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Agent string

const (
	Architect  Agent = "Architect"
	Auditor    Agent = "Auditor"
	Builder    Agent = "Builder"
	Strategist Agent = "Strategist"
	Sentinel   Agent = "Sentinel"
)

var (
	globalLogger *zap.Logger
	loggerOnce   sync.Once
)

// GetLogger returns the global zap logger, initializing it if necessary
func GetLogger() *zap.Logger {
	loggerOnce.Do(func() {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout", "SESSION_LOG.json"}
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		var err error
		globalLogger, err = config.Build()
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v\n", err)
			// Fallback to a basic logger if configuration fails
			globalLogger = zap.NewExample()
		}
	})
	return globalLogger
}

// LogAction logs an action with basic metadata
func LogAction(agent Agent, action, status, metadata string) error {
	GetLogger().Info(action,
		zap.String("agent", string(agent)),
		zap.String("status", status),
		zap.String("metadata", metadata),
	)
	return nil
}

// LogFullAction logs an action with full metadata including latency and tokens
func LogFullAction(agent Agent, action, status, metadata string, latency int64, tokens int) error {
	GetLogger().Info(action,
		zap.String("agent", string(agent)),
		zap.String("status", status),
		zap.String("metadata", metadata),
		zap.Int64("latency_ms", latency),
		zap.Int("tokens", tokens),
	)
	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
