package tui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/nycruz/gail/internal/assistant"
)

type LLM interface {
	Prompt(ctx context.Context, roleName string, rolePersona string, skillInstruction string, message string) (string, error)
	GetModel() string
	GetUser() string
}

// Interface Guard for Model
// Ensure Model implements tea.Model
var _ tea.Model = (*model)(nil)

type model struct {
	viewport        viewport.Model // Viewport for displaying chat conversation
	textarea        textarea.Model // Textarea for user input
	textAreaContent string         // Content of the textarea
	messagesDisplay []string       // Messages to display in viewport
	spinner         spinner.Model  // Spinner for loading state
	isLoading       bool           // Loading state
	senderStyle     lipgloss.Style // Style for user messages
	receiverStyle   lipgloss.Style // Style for Gail's messages
	helpSection     string         // Help section
	focusOnTextArea bool           // Focus on textarea

	statusBarMessage string

	assistant    *assistant.Assistant // Assistant
	isRolePrompt bool                 // Role prompt state
	roleList     list.Model           // List for displaying roles
	role         assistant.Role       // Current Role

	isSkillPrompt bool            // Skill prompt state
	skillList     list.Model      // List for displaying skills
	skill         assistant.Skill // Current Skill

	llm LLM // Large Language Model

	viewportCurrentWidth  int // Current width of the window
	viewportCurrentHeight int // Current height of the window
	textAreaCurrentWidth  int // Current width of the window
	textAreaCurrentHeight int // Current height of the window

	logger *slog.Logger // Logger
	err    error        // Error state
}

type errMsg error
type clearStatusBarMsg struct{}

const (
	clearStatusBarAfterSeconds time.Duration = 10
	defaultStatusMessage       string        = "'ctrl-q':quit, 'ctrl+s':send, 'ctrl+r':pick role, 'ctrl+e':pick skill, 'ctrl+d':save conversation, 'ctrl+c':copy conversation"
)

func New(logger *slog.Logger, mdl LLM, assistant *assistant.Assistant) model {
	ta := setupTextArea()
	vp := setupViewPort()
	s := setupSpinner()
	h := fadedStyle.Render("Type here...")

	roles := setupRoles(assistant.Roles)
	defaultRole := assistant.DefaultRole()

	skills := setupSkills(assistant.Skills)
	defaultSkill := assistant.DefaultSkill()

	return model{
		textarea:         ta,
		viewport:         vp,
		spinner:          s,
		isLoading:        false,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("6")),
		helpSection:      h,
		focusOnTextArea:  true,
		statusBarMessage: defaultStatusMessage,
		messagesDisplay:  []string{},
		assistant:        assistant,
		roleList:         roles,
		isRolePrompt:     false,
		role:             defaultRole,
		skillList:        skills,
		isSkillPrompt:    false,
		skill:            defaultSkill,
		llm:              mdl,
		logger:           logger,
		err:              nil,
	}
}

// Init
func (m model) Init() tea.Cmd {
	return m.textarea.Focus()
}

func (m model) View() string {
	if err := m.err; err != nil {
		m.statusBarMessage = fmt.Sprintf("Error: %v", err)
	}

	if m.isRolePrompt {
		return roleStyle.Render(m.roleList.View())
	}

	if m.isSkillPrompt {
		return skillStyle.Render(m.skillList.View())
	}

	if m.focusOnTextArea {
		textAreaStyle = textAreaStyle.BorderForeground(lipgloss.Color(TextHighlightColor))
		viewPortStyle = viewPortStyle.BorderForeground(lipgloss.Color(BorderColor))
	} else {
		viewPortStyle = viewPortStyle.BorderForeground(lipgloss.Color(TextHighlightColor))
		textAreaStyle = textAreaStyle.BorderForeground(lipgloss.Color(BorderColor))
	}

	if m.isLoading {
		m.statusBarMessage = fmt.Sprintf("%s thinking...", m.spinner.View())
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		m.viewportHeaderView(),
		viewPortStyle.Render(m.viewport.View()),
		m.viewPortFooterView(),
		m.textAreaHeaderView(),
		textAreaStyle.Render(m.textarea.View()),
		statusBarStyle.Render(m.statusBarMessage),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		sCmd  tea.Cmd
		rlCmd tea.Cmd
		slCmd tea.Cmd
	)

	// First, update the textarea
	m.textarea, tiCmd = m.textarea.Update(msg)

	// Conditionally update the viewport only if the textarea is not focused
	if !m.focusOnTextArea {
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	m.roleList, rlCmd = m.roleList.Update(msg)
	m.skillList, slCmd = m.skillList.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Do nothing when "q" is pressed to prevent quitting
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlQ:
			return m, tea.Quit

		// Ctrl+Shift+Tab to toggle focus between textarea and viewport
		case tea.KeyShiftTab:
			m.focusOnTextArea = !m.focusOnTextArea
			if m.focusOnTextArea {
				m.viewport.KeyMap.Up.SetEnabled(false)
				m.viewport.KeyMap.Down.SetEnabled(false)
				m.viewport.KeyMap.PageUp.SetEnabled(false)
				m.viewport.KeyMap.PageDown.SetEnabled(false)
				m.textarea.Focus()
			} else {
				m.viewport.KeyMap.Up.SetEnabled(true)
				m.viewport.KeyMap.Down.SetEnabled(true)
				m.viewport.KeyMap.PageUp.SetEnabled(true)
				m.viewport.KeyMap.PageDown.SetEnabled(true)
				m.textarea.Blur()
			}
			return m, nil

		// Ctrl+S to send the message
		case tea.KeyCtrlS:
			m.textAreaContent = m.textarea.Value()
			m.textarea.Reset()
			m.textarea.Blur()
			m.focusOnTextArea = false
			m.isLoading = true
			return m, tea.Batch(
				m.spinner.Tick,
				m.fetchAnswer(m.role.Name, m.role.Persona, m.skill.Instruction, m.textAreaContent),
			)

		// Ctrl+R to pick a role
		case tea.KeyCtrlR:
			m.isRolePrompt = true
			m.textarea.Blur()
			m.focusOnTextArea = false

		// Ctrl+E to pick a skill
		case tea.KeyCtrlE:
			m.isSkillPrompt = true
			m.textarea.Blur()
			m.focusOnTextArea = false

			skills := m.assistant.GetRoleSkills(m.role.ID)
			skillItems := skillList(skills)

			return m, m.skillList.SetItems(skillItems)

		// Ctrl+C to enter copy mode
		case tea.KeyCtrlC:
			unformmatedAnswer := removeANSICodes(strings.Join(m.messagesDisplay, "\n"))
			return m, m.copyModeRun(unformmatedAnswer)

		// Ctrl+D to save the conversation
		case tea.KeyCtrlD:
			unformmatedAnswer := removeANSICodes(strings.Join(m.messagesDisplay, "\n"))
			return m, m.saveConversation(unformmatedAnswer)

		// Ctrl+Enter to select a role or skill when in Select mode
		case tea.KeyEnter:
			if m.isRolePrompt {
				c, ok := m.roleList.SelectedItem().(RoleItem)
				if !ok {
					roleSelectError := "internal error: could not select Role"
					m.logger.Info(roleSelectError)
					m.err = errors.New(roleSelectError)
					return m, nil
				}

				m.role.Persona = c.FilterValue()
				m.role.ID = c.ID()
				m.isRolePrompt = false
				m.focusOnTextArea = true
				m.textarea.Focus()
			}
			if m.isSkillPrompt {
				c, ok := m.skillList.SelectedItem().(SkillItem)
				if !ok {
					skillSelectError := "internal error: could not select Skill"
					m.logger.Info(skillSelectError)
					m.err = errors.New(skillSelectError)
					return m, nil
				}

				m.skill.Instruction = c.FilterValue()
				m.isSkillPrompt = false
				m.focusOnTextArea = true
				m.textarea.Focus()
			}
		}

	case copyModeFinishedMsg:
		if msg.err != nil {
			m.statusBarMessage = fmt.Sprintf("Error entering copy mode: %v", msg.err)
		} else {
			m.statusBarMessage = msg.msg
		}
		return m, clearStatusBarAfter(clearStatusBarAfterSeconds * time.Second)

	case saveModeFinishedMsg:
		if msg.err != nil {
			m.statusBarMessage = fmt.Sprintf("Error saving conversation: %v", msg.err)
		} else {
			m.statusBarMessage = msg.msg
		}
		return m, clearStatusBarAfter(clearStatusBarAfterSeconds * time.Second)

	case Answer:
		if msg.Error != nil {
			m.statusBarMessage = fmt.Sprintf("Error fetching answer: %v", msg.Error)
		} else {
			m.statusBarMessage = msg.msg
		}

		userPrompt := m.senderStyle.Render("You: ") + m.textAreaContent
		userPrompt = wordwrap.String(userPrompt, m.viewportCurrentWidth-ReducerWidthForBorder)

		gailPrompt := m.receiverStyle.Render("\nGail: ") + msg.Answer + "\n"
		gailPrompt = wordwrap.String(gailPrompt, m.viewportCurrentWidth-ReducerWidthForBorder)

		m.messagesDisplay = append(m.messagesDisplay, userPrompt, gailPrompt)

		m.viewport.SetContent(strings.Join(m.messagesDisplay, "\n"))
		m.viewport.GotoBottom()
		m.isLoading = false

		unformmatedAnswer := removeANSICodes(strings.Join(m.messagesDisplay, "\n"))
		return m, tea.Batch(m.saveConversation(unformmatedAnswer), clearStatusBarAfter(clearStatusBarAfterSeconds*time.Second))

	// Clear the status bar when the timer expires
	case clearStatusBarMsg:
		m.statusBarMessage = defaultStatusMessage
		return m, nil

	case tea.WindowSizeMsg:
		widthWithBorder := msg.Width - ReducerWidth

		// viewport sizes
		viewportHeightWithBorder := msg.Height - int(float32(msg.Height)*ViewPortHeightPercentage) - ViewPortReducerHeight

		m.viewportCurrentWidth = widthWithBorder
		m.viewportCurrentHeight = viewportHeightWithBorder

		viewPortStyle.Width(m.viewportCurrentWidth)
		viewPortStyle.Height(m.viewportCurrentHeight)

		// reduce the width of the viewport to account for the border
		m.viewport.Width = m.viewportCurrentWidth - ReducerWidthForBorder
		m.viewport.Height = m.viewportCurrentHeight

		// textarea sizes
		textareaHeightWithBorder := msg.Height - int(float32(msg.Height)*TextAreaHeightPercentage) - TextAreaReducerHeight

		m.textAreaCurrentWidth = widthWithBorder
		m.textAreaCurrentHeight = textareaHeightWithBorder

		textAreaStyle.Width(m.textAreaCurrentWidth)
		textAreaStyle.Height(m.textAreaCurrentHeight)

		// reduce the width of the textarea to account for the border
		m.textarea.SetWidth(m.textAreaCurrentWidth - ReducerWidthForBorder + 2) // TODO: fix hardcoded number
		m.textarea.SetHeight(m.textAreaCurrentHeight)

		// status bar sizes
		statusBarStyle.Width(m.textAreaCurrentWidth)

		// role list sizes
		roleStyle.Width(msg.Width - ReducerWidth)
		roleStyle.Height(viewportHeightWithBorder)

		m.roleList.SetWidth(msg.Width - ReducerWidth)
		m.roleList.SetHeight(msg.Height - ReducerWidth)

		// skill list sizes
		skillStyle.Width(msg.Width - ReducerWidth)
		skillStyle.Height(viewportHeightWithBorder)

		m.skillList.SetWidth(msg.Width - ReducerWidth)
		m.skillList.SetHeight(msg.Height - ReducerWidth)

	case spinner.TickMsg:
		m.spinner, sCmd = m.spinner.Update(msg)

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd, sCmd, rlCmd, slCmd)
}

func setupTextArea() textarea.Model {
	ta := textarea.New()
	ta.Placeholder = "Type here..."
	ta.CharLimit = 0
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)
	ta.KeyMap.Paste.SetEnabled(true)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Focus()
	return ta
}

func setupViewPort() viewport.Model {
	v := viewport.New(0, 0)
	return v
}

func setupSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

// setupRoles creates a list.Model of roles for user to select from
func setupRoles(roles []assistant.Role) list.Model {
	roleItems := roleList(roles)

	rl := list.New(roleItems, list.NewDefaultDelegate(), 0, 0)
	rl.Title = "Roles"
	rl.SetShowHelp(true)
	rl.SetFilteringEnabled(true)

	return rl
}

func roleList(roles []assistant.Role) []list.Item {
	roleItems := []list.Item{}
	for _, role := range roles {
		roleItems = append(roleItems, RoleItem{
			id:      string(role.ID),
			name:    role.Name,
			persona: role.Persona,
		})
	}

	return roleItems
}

// setupSkills creates a list.Model of skills for user to select from
func setupSkills(skills []assistant.Skill) list.Model {
	skillItems := skillList(skills)

	sl := list.New(skillItems, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Skills"
	sl.SetShowHelp(true)
	sl.SetFilteringEnabled(true)

	return sl
}

func skillList(skills []assistant.Skill) []list.Item {
	skillItems := []list.Item{}
	for _, skill := range skills {
		skillItems = append(skillItems, SkillItem{
			id:          skill.ID,
			instruction: skill.Instruction,
			description: skill.Description,
		})
	}

	return skillItems
}

func removeANSICodes(input string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansi.ReplaceAllString(input, "")
}

func clearStatusBarAfter(d time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(d)
		return clearStatusBarMsg{}
	}
}
