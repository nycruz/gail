package tui

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type copyModeFinishedMsg struct {
	msg string
	err error
}

func (m model) copyModeRun(content string) tea.Cmd {
	tmpfile, err := os.CreateTemp("", "tui-*.md")
	if err != nil {
		m.logger.Info(fmt.Sprintf("error creating temp file: %s", err))
		log.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		m.logger.Info(fmt.Sprintf("error writing to temp file: %s", err))
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		m.logger.Info(fmt.Sprintf("error closing temp file: %s", err))
		log.Fatal(err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		defer os.Remove(tmpfile.Name())
		if err != nil {
			m.logger.Info(fmt.Sprintf("error running editor: %s", err))
			return copyModeFinishedMsg{err: err}
		}

		return copyModeFinishedMsg{msg: "Entered copy mode"}
	})
}
