package connect4ai

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"websockets/ai"
	ctypes "websockets/games/connect4/types"
	"websockets/games/connect4/types/internalstate"
)

type State struct {
	state *ctypes.UpdateGameState
}

func (state *State) UnmarshalJSON(stateJson []byte) error {
	return json.Unmarshal(stateJson, state.state)
}

type Action struct {
	Col     int
	Rematch bool
}

func (action *Action) MarshalJSON() ([]byte, error) {
	tom := map[string]interface{}{"Col": action.Col, "Rematch": action.Rematch}
	return json.Marshal(tom)
}

type Agent struct {
	Nodes       map[string]*Node
	AgentId     string
	RematchSent bool
}

func NewAgent(id string) *Agent {
	return &Agent{
		AgentId: id,
		Nodes:   map[string]*Node{},
	}
}

type Node struct {
	color    int
	move     int
	wins     int
	total    int
	children map[int]*Node
	parent   *Node
}

func (agent *Agent) Selection(node *Node, is *internalstate.InternalState) (*Node, int) {
	current := node
	for {
		moves := is.GenerateMoves()
		if len(moves) == 0 {
			return nil, -1
		}
		action := moves[rand.Intn(len(moves))]
		is.MakeMove(action)
		if child, ok := current.children[action]; ok {
			current = child
		} else {
			return current, action
		}
	}
	return nil, -1
}

func (agent *Agent) Expand(node *Node, is *internalstate.InternalState, action int) (string, *Node) {
	child_state := is.ToString()
	child := &Node{
		color:    1 - is.Turn,
		move:     action,
		children: map[int]*Node{},
		parent:   node,
	}
	node.children[action] = child
	return child_state, child
}

func (agent *Agent) Simulation(node *Node, is *internalstate.InternalState) int {
	for {
		if is.StalemateCheck() {
			return -1
		} else {
			vc := is.VictoryCheck()
			if vc >= 0 {
				return vc
			}
		}
		legalMoves := is.GenerateMoves()
		nextMove := legalMoves[rand.Intn(len(legalMoves))]
		is.MakeMove(nextMove)
	}
	return -1
}

func (agent *Agent) Backpropagation(node *Node, winner int) {
	for current := node; current != nil; current = current.parent {
		current.total += 1
		if int(current.color) == winner {
			current.wins += 1
		}
	}
}

func (agent *Agent) stateReset(moveCount int, is *internalstate.InternalState) {
	for len(is.Moves) > moveCount {
		is.UnmakeMove()
	}
}

func (agent *Agent) Search(current *Node, is *internalstate.InternalState) {
	defer agent.stateReset(len(is.Moves), is)

	node, action := agent.Selection(current, is)
	if node == nil {
		return
	}

	key, child := agent.Expand(node, is, action)
	agent.Nodes[key] = child

	winner := agent.Simulation(child, is)

	agent.Backpropagation(child, winner)
}

func (agent *Agent) RunSearch(duration time.Duration, is *internalstate.InternalState) int {
	timer := time.After(duration)

	current := agent.InitialNode(is)
	if current == nil {
		panic("INITIAL NODE RETURNED NIL")
	}
	cont := true
	for cont {
		select {
		case <-timer:
			cont = false
		default:
		}
		agent.Search(current, is)
	}

	best_ratio := 0.0
	toret := -1

	for action, child := range current.children {
		fmt.Printf("Child: %+v\n", *child)
		win_ratio := float64(child.wins) / float64(child.total)
		if win_ratio > best_ratio {
			toret = action
			best_ratio = win_ratio
		}
	}
	if toret < 0 {
		moves := is.GenerateMoves()
		return moves[rand.Intn(len(moves))]
	}
	return toret
}

func (agent *Agent) InitialNode(is *internalstate.InternalState) *Node {
	state_str := is.ToString()
	if _, ok := agent.Nodes[state_str]; !ok {
		agent.Nodes[state_str] = &Node{
			children: map[int]*Node{},
			color:    -1,
			move:     -1,
			wins:     -1,
			total:    -1,
		}
	}
	return agent.Nodes[state_str]
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

func (agent *Agent) GenerateAction(state ai.State) ai.Action {
	s := state.(*State)
	is := internalstate.NewInternalState(agent.AgentId, s.state)
	fresh := internalstate.NewInternalState(agent.AgentId, s.state)
	actions := state.LegalActions()
	a := actions[0].(*Action)
	agent.RematchSent = a.Rematch
	if a.Rematch {
		return a
	}
	if len(actions) == 1 {
		return actions[0].(*Action)
	}
	action := agent.RunSearch(1000*time.Millisecond, is)
	fmt.Println("Action:", action)
	fmt.Println(fresh.ToString())
	fmt.Println(is.ToString())
	return &Action{
		Col: action,
	}
}
