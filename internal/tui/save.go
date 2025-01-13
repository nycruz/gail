package tui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type saveModeFinishedMsg struct {
	err error
}

func (m model) saveConversation(content string) {
	homeDir, _ := os.UserHomeDir()
	historyDir := filepath.Join(homeDir, "gail_history")
	err := os.MkdirAll(historyDir, os.ModePerm)
	if err != nil {
		m.logger.Info(fmt.Sprintf("error creating directory in '%s': %s", historyDir, err))
		log.Fatal(err)
	}

	now := time.Now()
	filename := fmt.Sprintf("conversation_%s.txt", now.Format("2006-01-02-15-04-05"))
	path := fmt.Sprintf("%s/%s", historyDir, filename)

	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		m.logger.Info(fmt.Sprintf("error writing to conversation file: %s", err))
		log.Fatal(err)
	}
}
