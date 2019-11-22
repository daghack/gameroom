package connect4ai

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"websockets/ai"
	ctypes "websockets/games/connect4/types"
	internalstate "websockets/games/connect4/types/internalstate"
)

type State struct {
	state *ctypes.UpdateGameState
}

func score(is *internalstate.InternalState) int {
	victor := is.VictoryCheck()
	if victor < 0 {
		if is.StalemateCheck() {
			return -100
		}
		return evaluate(is)
	}
	if victor == is.Agent {
		return 1000
	}
	return -1000
}

func evaluate(is *internalstate.InternalState) int {
	return rand.Intn(100)
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
		if len(col) < ctypes.Height {
			toret = append(toret, &Action{
				Col: i,
			})
		}
	}
	if len(toret) == 0 {
		toret = append(toret, &Action{
			Rematch: true,
		})
		return toret
	}
	return toret
}

func (agent *Agent) BaseState() ai.State {
	return &State{
		state: &ctypes.UpdateGameState{},
	}
}

func (agent *Agent) CanAct(state ai.State) bool {
	s := state.(*State)
	correctTurn := s.state.Players[agent.AgentId] == s.state.CurrentTurn && !s.state.GameOver
	rematch := s.state.GameOver && !agent.RematchSent
	return correctTurn || rematch
}

func (agent *Agent) min(is *internalstate.InternalState, alpha, beta, p_action, depth int) (int, int) {
	if depth == 0 {
		return p_action, score(is)
	}
	actions := is.GenerateMoves()

	bestAction := actions[0]
	bestScore := 100000

	for _, action := range actions {
		is.MakeMove(action)
		_, score := agent.max(is, alpha, beta, action, depth-1)
		is.UnmakeMove(action)
		if score < bestScore {
			bestAction = action
			bestScore = score
			if bestScore < beta {
				beta = bestScore
			}
			if alpha >= beta {
				break
			}
		}
	}
	return bestAction, bestScore
}

func (agent *Agent) max(is *internalstate.InternalState, alpha, beta, p_action, depth int) (int, int) {
	if depth == 0 || is.StalemateCheck() || is.VictoryCheck() >= 0 {
		return p_action, score(is)
	}
	actions := is.GenerateMoves()

	bestAction := actions[0]
	bestScore := -100000

	for _, action := range actions {
		is.MakeMove(action)
		_, score := agent.min(is, alpha, beta, action, depth-1)
		is.UnmakeMove(action)
		if score > bestScore {
			bestAction = action
			bestScore = score
			if bestScore > alpha {
				alpha = bestScore
			}
			if alpha >= beta {
				break
			}
		}
	}
	return bestAction, bestScore
}

func (agent *Agent) GenerateAction(state ai.State) ai.Action {
	s := state.(*State)
	is := internalstate.NewInternalState(agent.AgentId, s.state)
	actions := state.LegalActions()
	a := actions[0].(*Action)
	agent.RematchSent = a.Rematch
	if a.Rematch {
		return a
	}
	action, score := agent.max(is, -10000000, 10000000, 0, 7)
	fmt.Println("Action: ", action)
	fmt.Println("Score: ", score)
	return &Action{
		Col: action,
	}
}
