package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	TextAreaHeightPercentage float32 = 0.8
	ViewPortHeightPercentage float32 = 0.2

	// ReducerWidth is the amount of characters to reduce so the borders do not touch the edges of the terminal window
	ReducerWidth                int = 2
	ReducerWidthForBorder       int = 2
	ViewPortContentReducerWidth int = 4
	TextAreaReducerHeight       int = 2
	// ViewPortReducerWidth is the amount of characters to reduce so the borders do not touch the edges of the terminal window
	ViewPortReducerHeight int = 3

	BoderColor = "8"
)

var (
	textAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color(BoderColor)).
			Margin(0, 0, 0, 0)

	viewPortStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, true).
			BorderForeground(lipgloss.Color(BoderColor)).
			Padding(0, 1).
			Margin(0, 0, 0, 0)

	fadedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(BoderColor))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Border(lipgloss.HiddenBorder()).
			Padding(0, 0).
			Margin(0, 0, 0, 0)

	roleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())

	skillStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())

	titleStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().Padding(0, 0).Margin(0, 0).Foreground(lipgloss.Color("6"))
	}()

	infoStyle = func() lipgloss.Style {
		return titleStyle.Copy().Foreground(lipgloss.Color("6"))
	}()
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) viewportHeaderView() string {
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BoderColor))
	if !m.focusOnTextArea {
		l.Foreground(lipgloss.Color("6"))
	}

	title := titleStyle.Render(fmt.Sprintf(" %s ", m.llm.GetUser()))
	titleWidth := lipgloss.Width(title)
	leftBorderWidth := (m.viewportCurrentWidth - titleWidth) / 2
	rightBorderWidth := m.viewportCurrentWidth - titleWidth - leftBorderWidth
	leftBorder := strings.Repeat("─", max(0, leftBorderWidth))
	rightBorder := strings.Repeat("─", max(0, rightBorderWidth))

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
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BoderColor))
	if !m.focusOnTextArea {
		l.Foreground(lipgloss.Color("6"))
	}

	modelName := infoStyle.Foreground(lipgloss.Color(BoderColor)).Render(m.llm.GetModel())
	scrollPercent := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	borderLines := strings.Repeat("─", max(0, m.viewportCurrentWidth-lipgloss.Width(scrollPercent)-lipgloss.Width(modelName)))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		l.Render("╰"),
		l.Render(modelName),
		l.Render(borderLines),
		scrollPercent,
		l.Render("╯"))
}

func (m model) textAreaHeaderView() string {
	l := lipgloss.NewStyle().Foreground(lipgloss.Color(BoderColor))
	if m.focusOnTextArea {
		l.Foreground(lipgloss.Color("6"))
	}
	title := titleStyle.Render(" Prompt ")
	titleWidth := lipgloss.Width(title)
	leftBorderWidth := (m.textAreaCurrentWidth - titleWidth) / 2
	rightBorderWidth := m.textAreaCurrentWidth - titleWidth - leftBorderWidth
	leftBorder := strings.Repeat("─", max(0, leftBorderWidth))
	rightBorder := strings.Repeat("─", max(0, rightBorderWidth))

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		l.Render("╭"),
		l.Render(leftBorder),
		title,
		l.Render(rightBorder),
		l.Render("╮"),
	)
}
