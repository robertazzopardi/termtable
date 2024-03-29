package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Action string

const (
	SUBMIT Action = "SUBMIT"
	TEST   Action = "TEST"
)

var (
	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	focusedTestButton = focusedStyle.Copy().Render("[ Test ]")
	blurredTestButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Test"))
)

type NewConnectionModel struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	valid      bool
	action     Action
}

func InitialNewConnectionModel() NewConnectionModel {
	var newConnectionInputs = []string{"Connection's Name", "Host", "Port", "User", "Password", "Database Name"}
	m := NewConnectionModel{
		inputs: make([]textinput.Model, len(newConnectionInputs)),
		action: SUBMIT,
	}

	var t textinput.Model
	for i, value := range newConnectionInputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32
		t.Placeholder = value

		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}

		if value == "Password" {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (m *NewConnectionModel) Validate() bool {
	for i, input := range m.inputs {
		if input.Value() == "" {
			// m.focusIndex = i - 1
			m.focusIndex = i
			input.TextStyle = focusedStyle
			return false
		}
	}
	return true
}

func (m NewConnectionModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m NewConnectionModel) Update(msg tea.Msg) (NewConnectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Handle button actions
		case "left", "right":
			if m.focusIndex == len(m.inputs) {
				if m.action == SUBMIT {
					m.action = TEST
				} else {
					m.action = SUBMIT
				}
			}

		case "enter":
			if m.focusIndex == len(m.inputs) {
				switch m.action {
				case SUBMIT:
					if m.Validate() {
						m.valid = true
						// return m, nil
					}
				case TEST:
					fmt.Println("Test")
				}

			}

			return m.updateInputStates()

		// Set focus to next input
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			return m.updateInputStates()
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *NewConnectionModel) updateInputStates() (NewConnectionModel, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}

	return *m, tea.Batch(cmds...)
}

func (m *NewConnectionModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m NewConnectionModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	submitButton := &blurredButton
	testButton := &blurredTestButton
	if m.focusIndex == len(m.inputs) {
		if m.action == SUBMIT {
			submitButton = &focusedButton
		} else {
			testButton = &focusedTestButton
		}
	}
	fmt.Fprintf(&b, "\n\n%s%s\n\n", *submitButton, *testButton)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
