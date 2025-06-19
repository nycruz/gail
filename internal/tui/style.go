package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	TextHighlightColor = getEnvWithDefault("GAIL_HIGHLIGHT_COLOR", "6")
)

const (
	TextAreaHeightPercentage float32 = 0.8
	ViewPortHeightPercentage float32 = 0.2

	// ReducerWidth is the amount of characters to reduce so the borders do not touch the edges of the terminal window
	ReducerWidth                int = 2
	ReducerWidthForBorder       int = 2
	ViewPortContentReducerWidth int = 4
	TextAreaReducerHeight       int = 1
	// ViewPortReducerWidth is the amount of characters to reduce so the borders do not touch the edges of the terminal window
	ViewPortReducerHeight int = 7

	BorderColor = "8"
)

var (
	textAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color(BorderColor)).
			Margin(0, 0, 0, 0)

	viewPortStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, true).
			BorderForeground(lipgloss.Color(BorderColor)).
			Padding(0, 1).
			Margin(0, 0, 0, 0)

	fadedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(BorderColor))

	// spinnerStyle = lipgloss.NewStyle().
	// 		Foreground(lipgloss.Color(TextHighlightColor)).
	// 		Border(lipgloss.HiddenBorder()).
	// 		Padding(0, 0).
	// 		Margin(0, 0, 0, 0)

	roleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())

	skillStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())

	statusBarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color(BorderColor)).
			Foreground(lipgloss.Color(TextHighlightColor)).
			Padding(0, 0, 0, 1).
			Margin(0, 0, 0, 0)

	titleStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Padding(0, 0).Margin(0, 0).Foreground(lipgloss.Color(TextHighlightColor))
	}()

	infoStyle = func() lipgloss.Style {
		ts := titleStyle
		return ts.Foreground(lipgloss.Color(TextHighlightColor))
	}()
)

func getMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) viewportHeaderView() string {
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BorderColor))
	if !m.focusOnTextArea {
		l = l.Foreground(lipgloss.Color(TextHighlightColor))
	}

	title := titleStyle.Render(fmt.Sprintf(" %s ", m.llm.GetUser()))
	titleWidth := lipgloss.Width(title)
	leftBorderWidth := (m.viewportCurrentWidth - titleWidth) / 2
	rightBorderWidth := m.viewportCurrentWidth - titleWidth - leftBorderWidth
	leftBorder := strings.Repeat("─", getMax(0, leftBorderWidth))
	rightBorder := strings.Repeat("─", getMax(0, rightBorderWidth))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		l.Render("╭"),
		l.Render(leftBorder),
		title,
		l.Render(rightBorder),
		l.Render("╮"),
	)
}

func (m model) viewPortFooterView() string {
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BorderColor))
	if !m.focusOnTextArea {
		l = l.Foreground(lipgloss.Color(TextHighlightColor))
	}

	modelName := infoStyle.Foreground(lipgloss.Color(BorderColor)).Render(m.llm.GetModel())
	scrollPercent := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	borderLines := strings.Repeat("─", getMax(0, m.viewportCurrentWidth-lipgloss.Width(scrollPercent)-lipgloss.Width(modelName)))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		l.Render("╰"),
		l.Render(modelName),
		l.Render(borderLines),
		scrollPercent,
		l.Render("╯"))
}

func (m model) textAreaHeaderView() string {
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BorderColor))
	if m.focusOnTextArea {
		l = l.Foreground(lipgloss.Color(TextHighlightColor))
	}
	title := titleStyle.Render(" Prompt ")
	titleWidth := lipgloss.Width(title)
	leftBorderWidth := (m.textAreaCurrentWidth - titleWidth) / 2
	rightBorderWidth := m.textAreaCurrentWidth - titleWidth - leftBorderWidth
	leftBorder := strings.Repeat("─", getMax(0, leftBorderWidth))
	rightBorder := strings.Repeat("─", getMax(0, rightBorderWidth))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		l.Render("╭"),
		l.Render(leftBorder),
		title,
		l.Render(rightBorder),
		l.Render("╮"),
	)
}

// getEnvWithDefault retrieves the value of the environment variable named by key.
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
