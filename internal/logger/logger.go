package logger

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

func New(level slog.Level) (*slog.Logger, error) {
	filename, err := setupLogFile()
	if err != nil {
		return nil, fmt.Errorf("failed to setup the log file: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	l := slog.New(slog.NewTextHandler(filename, opts))
	return l, nil
}

// setupLogFile creates a log file in the user's default log directory
func setupLogFile() (*os.File, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	logDirPath := fmt.Sprintf("%s/Library/Logs/gail", homeDir)
	if _, err := os.Stat(logDirPath); os.IsNotExist(err) {
		err = os.Mkdir(logDirPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	logFileName := fmt.Sprintf("gail_%s.log", time.Now().Format("2006-01-02"))
	logsDir := fmt.Sprintf("%s/%s", logDirPath, logFileName)
	logFile, err := os.OpenFile(logsDir, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return logFile, nil
}
