package connect4ai

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"websockets/ai"
	"websockets/games/connect4"
)

type State struct {
	state *connect4.UpdateGameState
}

type Action struct {
	Col     int
	Rematch bool
}

type Agent struct {
	AgentId     string
	RematchSent bool
}

func (action *Action) MarshalJSON() ([]byte, error) {
	tom := map[string]interface{}{"Col": action.Col, "Rematch": action.Rematch}
	return json.Marshal(tom)
}

func (state *State) UnmarshalJSON(stateJson []byte) error {
	return json.Unmarshal(stateJson, state.state)
}

func (state *State) LegalActions() []ai.Action {
	toret := []ai.Action{}
	if state.state.GameOver {
		toret = append(toret, &Action{
			Rematch: true,
		})
		return toret
	}
	for i, col := range state.state.Columns {
		if len(col) < connect4.Height {
			toret = append(toret, &Action{
				Col: i,
			})
		}
	}
	return toret
}

func (agent *Agent) BaseState() ai.State {
	return &State{
		state: &connect4.UpdateGameState{},
	}
}

func (agent *Agent) CanAct(state ai.State) bool {
	s := state.(*State)
	fmt.Printf("%+v\n", *s.state)
	correctTurn := s.state.Players[agent.AgentId] == s.state.CurrentTurn && !s.state.GameOver
	rematch := s.state.GameOver && !agent.RematchSent
	return correctTurn || rematch
}

func (agent *Agent) GenerateAction(state ai.State) ai.Action {
	legalActions := state.LegalActions()
	action := legalActions[0].(*Action)
	agent.RematchSent = action.Rematch
	return legalActions[rand.Intn(len(legalActions))]
}
