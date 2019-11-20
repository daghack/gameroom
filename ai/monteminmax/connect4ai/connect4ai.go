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

type internalState struct {
	board  [][]int
	height []int
	turn   int
	agent  int
}

func buildInternalState(agentId string, s *connect4.UpdateGameState) *internalState {
	toret := &internalState{
		board:  [][]int{},
		height: []int{},
		turn:   int(s.CurrentTurn),
		agent:  int(s.Players[agentId]),
	}
	for col := 0; col < connect4.Width; col += 1 {
		toret.board = append(toret.board, []int{})
		toret.height = append(toret.height, len(s.Columns[col]))
		for row := 0; row < connect4.Height; row += 1 {
			if row < len(s.Columns[col]) {
				toret.board[col] = append(toret.board[col], int(s.Columns[col][row]))
			} else {
				toret.board[col] = append(toret.board[col], -1)
			}
		}
	}
	return toret
}

func (is *internalState) score() int {
	victor := is.victoryCheck()
	if victor < 0 {
		if is.stalemateCheck() {
			return -10000
		}
		return is.evaluate()
	}
	if victor == is.agent {
		return 100000
	}
	return -100000
}

func (is *internalState) rollout() int {
	victor := is.victoryCheck()
	if victor > -1 {
		if victor == is.agent {
			return 1
		} else {
			return 0
		}
	}
	moves := is.generateMoves()
	if len(moves) == 0 {
		return 0
	}
	rand := moves[rand.Intn(len(moves))]
	is.makeMove(rand)
	defer is.unmakeMove(rand)
	return is.rollout()
}

func (is *internalState) evaluate() int {
	payout := 0
	for i := 0; i < 1000; i += 1 {
		payout += is.rollout()
	}
	return payout
}

func (is *internalState) generateMoves() []int {
	toret := []int{}
	for i, h := range is.height {
		if h < connect4.Height {
			toret = append(toret, i)
		}
	}
	return toret
}

func (is *internalState) makeMove(col int) {
	height := is.height[col]
	is.board[col][height] = is.turn
	is.height[col] = height + 1
	is.turn = 1 - is.turn
}

func (is *internalState) unmakeMove(col int) {
	height := is.height[col]
	is.board[col][height-1] = -1
	is.height[col] = height - 1
	is.turn = 1 - is.turn
}

func (is *internalState) stalemateCheck() bool {
	for _, h := range is.height {
		if h != connect4.Height {
			return false
		}
	}
	return true
}

func (is *internalState) victoryCheck() int {
	v := is.colVictoryCheck()
	if v > -1 {
		return v
	}
	v = is.rowVictoryCheck()
	if v > -1 {
		return v
	}
	v = is.diagTopRightCheck()
	if v > -1 {
		return v
	}
	return is.diagTopLeftCheck()
}

func (is *internalState) colVictoryCheck() int {
	for col, h := range is.height {
		if h < 4 {
			continue
		}
		match := true
		top := is.board[col][h-1]
		for i := 0; i < 3; i += 1 {
			match = match && top == is.board[col][h-2-i]
		}
		if match {
			return top
		}
	}
	return -1
}

func (is *internalState) rowVictoryCheck() int {
	checkfor := is.height[:connect4.Width-3]
	for col, h := range checkfor {
		if h == 0 {
			continue
		}
		for row := 0; row < h; row += 1 {
			match := true
			top := is.board[col][row]
			for i := 0; i < 3; i += 1 {
				match = match && top == is.board[col+1+i][row]
			}
			if match {
				return top
			}
		}
	}
	return -1
}

func (is *internalState) diagTopRightCheck() int {
	checkfor := is.height[:connect4.Width-3]
	for col, h := range checkfor {
		if h > connect4.Height-3 {
			h = connect4.Height - 3
		}
		for row := 0; row < h; row += 1 {
			match := true
			top := is.board[col][row]
			for i := 0; i < 3; i += 1 {
				match = match && top == is.board[col+1+i][row+1+i]
			}
			if match {
				return top
			}
		}
	}
	return -1
}

func (is *internalState) diagTopLeftCheck() int {
	checkfor := is.height[connect4.Width-4:]
	for col, h := range checkfor {
		if h > connect4.Height-3 {
			h = connect4.Height - 3
		}
		col = col + 3
		for row := 0; row < h; row += 1 {
			match := true
			top := is.board[col][row]
			for i := 0; i < 3; i += 1 {
				match = match && top == is.board[col-1-i][row+1+i]
			}
			if match {
				return top
			}
		}
	}
	return -1
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
		state: &connect4.UpdateGameState{},
	}
}

func (agent *Agent) CanAct(state ai.State) bool {
	s := state.(*State)
	correctTurn := s.state.Players[agent.AgentId] == s.state.CurrentTurn && !s.state.GameOver
	rematch := s.state.GameOver && !agent.RematchSent
	return correctTurn || rematch
}

func (agent *Agent) min(is *internalState, alpha, beta, p_action, depth int) (int, int) {
	if depth == 0 {
		return p_action, is.score()
	}
	actions := is.generateMoves()

	bestAction := actions[0]
	bestScore := 100000

	for _, action := range actions {
		is.makeMove(action)
		_, score := agent.max(is, alpha, beta, action, depth-1)
		is.unmakeMove(action)
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

func (agent *Agent) max(is *internalState, alpha, beta, p_action, depth int) (int, int) {
	if depth == 0 || is.stalemateCheck() || is.victoryCheck() >= 0 {
		return p_action, is.score()
	}
	actions := is.generateMoves()

	bestAction := actions[0]
	bestScore := -100000

	for _, action := range actions {
		is.makeMove(action)
		_, score := agent.min(is, alpha, beta, action, depth-1)
		is.unmakeMove(action)
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
	is := buildInternalState(agent.AgentId, s.state)
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
