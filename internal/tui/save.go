package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

		filename := createFileName(content)
		path := fmt.Sprintf("%s/%s", historyDir, filename)
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return saveModeFinishedMsg{err: err}
		}

		return saveModeFinishedMsg{msg: fmt.Sprintf("Conversation saved in %s", path)}
	}
}

func createFileName(content string) string {
	contentFirstLine := strings.Split(content, "\n")[0]
	fileTitle := contentFirstLine
	if len(contentFirstLine) > 30 {
		fileTitle = contentFirstLine[:30]
	}

	fileTitle = strings.ReplaceAll(fileTitle, "You:", "")
	fileTitle = strings.ReplaceAll(fileTitle, " ", "_")

	now := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("%s_%s.md", now, fileTitle)

	return filename
}
