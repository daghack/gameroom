package ai

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type State interface {
	json.Unmarshaler
	LegalActions() []Action
}

type Action interface {
	json.Marshaler
}

type Agent interface {
	CanAct(State) bool
	GenerateAction(State) Action
	BaseState() State
}

type AI struct {
	websocketURL    string
	state           State
	stateUpdateChan chan State
	actionSendChan  chan Action
	agent           Agent
	conn            *websocket.Conn
}

func NewAgent(agent Agent, websocketURL string) (*AI, error) {
	toret := &AI{
		websocketURL: websocketURL,
		agent:        agent,
		state:        agent.BaseState(),
	}
	return toret, nil
}

func (ai *AI) startActionWriteLoop() {
	actionChan := make(chan Action, 8)
	ai.actionSendChan = actionChan
	go func() {
		for action := range ai.actionSendChan {
			actionJson, err := action.MarshalJSON()
			if err != nil {
				fmt.Println("WARNING:", err.Error())
			}
			err = ai.conn.WriteMessage(websocket.TextMessage, actionJson)
			if err != nil {
				// Can't write, connection is probably closed
				return
			}
		}
	}()
}

func (ai *AI) Close() {
	close(ai.stateUpdateChan)
	close(ai.actionSendChan)
	ai.conn.Close()
}

func (ai *AI) startStateReadLoop() {
	stateChan := make(chan State, 8)
	ai.stateUpdateChan = stateChan
	go func() {
		for {
			mtype, msg, err := ai.conn.ReadMessage()
			if err != nil {
				ai.Close()
				return
			}
			switch mtype {
			case websocket.TextMessage:
				ai.state.UnmarshalJSON(msg)
				ai.stateUpdateChan <- ai.state
			default:
				continue
			}
		}
	}()
}

func (ai *AI) Run() error {
	ws, _, err := websocket.DefaultDialer.Dial(ai.websocketURL, nil)
	if err != nil {
		return err
	}
	ai.conn = ws
	ai.startActionWriteLoop()
	ai.startStateReadLoop()
	for state := range ai.stateUpdateChan {
		if ai.agent.CanAct(state) {
			fmt.Println("AGENT CAN ACT")
			action := ai.agent.GenerateAction(state)
			ai.actionSendChan <- action
		} else {
			fmt.Println("AGENT CANNOT ACT")
		}
	}
	return nil
}
