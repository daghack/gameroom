package connect4

import ()

const (
	Red = iota
	Black
)

const (
	Width  int = 7
	Height int = 6
)

type Color int

type Position struct {
	Row int
	Col int
}

type MoveData struct {
	Col     int
	Rematch bool
}

type GameState struct {
	CurrentTurn      Color
	Columns          [Width][]Color
	GameOver         bool
	WinningPositions []Position
}

type UpdateGameState struct {
	GameState
	Players map[string]Color
}

type PlayerInfo struct {
	PlayerColor    Color
	UpdateChan     chan []byte
	RematchAttempt bool
}

func NewGameState() GameState {
	return GameState{
		Columns: [Width][]Color{{}, {}, {}, {}, {}, {}, {}},
	}
}
