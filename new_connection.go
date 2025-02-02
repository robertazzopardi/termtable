package main

type Action string

const (
	SUBMIT Action = "SUBMIT"
	TEST   Action = "TEST"
	CANCEL Action = "CANCEL"
)

type TestStatus string

const (
	PASSED TestStatus = "PASSED"
	FAILED TestStatus = "FAILED"
)

type NewConnectionModel struct {
	connection Connection
	testStatus TestStatus
	action     Action
}

// func (m NewConnectionModel) Update(msg tea.Msg) (NewConnectionModel, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {

// 		case "enter":
// 			if m.focusIndex == len(m.inputs) {
// 				conn := Connection{
// 					Host:     m.inputs[0].Value(),
// 					Port:     m.inputs[1].Value(),
// 					User:     m.inputs[2].Value(),
// 					Password: m.inputs[3].Value(),
// 					Database: m.inputs[4].Value(),
// 					Name:     m.inputs[5].Value(),
// 					status:   DISCONNECTED,
// 				}

// 				switch m.action {
// 				case SUBMIT:
// 					if conn.TestConnection() == PASSED {
// 						m.connection = conn
// 					}
// 				case TEST:
// 					if m.testStatus == NA {
// 						m.testStatus = conn.TestConnection()
// 					}

// 				}

// 			}

// }
