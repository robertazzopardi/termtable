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

type TestStatus string

const (
	PASSED TestStatus = "PASSED"
	FAILED TestStatus = "FAILED"
	NA     TestStatus = "NA"
)

var (
	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	focusedTestButton = focusedStyle.Copy().Render("[ Test ]")
	blurredTestButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Test"))
	errorTestButton   = fmt.Sprintf("[ %s ]", errorStyle.Render("Test"))
	successTestButton = fmt.Sprintf("[ %s ]", successStyle.Render("Test"))
)

type NewConnectionModel struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	connection Connection
	testStatus TestStatus
	action     Action
}

func InitialNewConnectionModel() NewConnectionModel {
	var newConnectionInputs = []string{
		"Host",
		"Port",
		"User",
		"Pass",
		"Name",
	}
	m := NewConnectionModel{
		inputs:     make([]textinput.Model, len(newConnectionInputs)),
		action:     SUBMIT,
		testStatus: NA,
	}

	var t textinput.Model
	for i, value := range newConnectionInputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32
		t.Placeholder = value

		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedItemStyle
			t.TextStyle = focusedItemStyle
		}

		if value == "Pass" {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
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

			if m.testStatus != NA {
				m.testStatus = NA
			}

		case "enter":
			if m.focusIndex == len(m.inputs) {
				conn := Connection{
					Host:   m.inputs[0].Value(),
					Port:   m.inputs[1].Value(),
					User:   m.inputs[2].Value(),
					Pass:   m.inputs[3].Value(),
					Name:   m.inputs[4].Value(),
					status: DISCONNECTED,
				}

				switch m.action {
				case SUBMIT:
					if conn.TestConnection() == PASSED {
						m.connection = conn
					}
				case TEST:
					if m.testStatus == NA {
						m.testStatus = conn.TestConnection()
					}

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
			m.inputs[i].PromptStyle = focusedItemStyle
			m.inputs[i].TextStyle = focusedItemStyle
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
		switch m.action {
		case SUBMIT:
			submitButton = &focusedButton
		case TEST:
			switch m.testStatus {
			case PASSED:
				testButton = &successTestButton
			case FAILED:
				testButton = &errorTestButton
			case NA:
				testButton = &focusedTestButton
			}
		}
	}
	fmt.Fprintf(&b, "\n\n%s%s\n\n", *submitButton, *testButton)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
