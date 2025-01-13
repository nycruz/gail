package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type saveModeFinishedMsg struct {
	msg string
	err errMsg
}

func (m model) saveConversation(content string) tea.Cmd {
	return func() tea.Msg {
		homeDir, _ := os.UserHomeDir()
		historyDir := filepath.Join(homeDir, "gail_history")
		err := os.MkdirAll(historyDir, os.ModePerm)
		if err != nil {
			return saveModeFinishedMsg{err: err}
		}

		now := time.Now()
		filename := fmt.Sprintf("conversation_%s.txt", now.Format("2006-01-02-15-04-05"))
		path := fmt.Sprintf("%s/%s", historyDir, filename)

		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return saveModeFinishedMsg{err: err}
		}

		return saveModeFinishedMsg{msg: fmt.Sprintf("Conversation saved to %s", path)}
	}
}
