package internalstate

import (
	ctypes "websockets/games/connect4/types"
)

type InternalState struct {
	board  [][]int
	height []int
	turn   int
	Agent  int
	moves  []int
}

func NewInternalState(agentId string, s *ctypes.UpdateGameState) *InternalState {
	toret := &InternalState{
		board:  [][]int{},
		height: []int{},
		turn:   int(s.CurrentTurn),
		Agent:  int(s.Players[agentId]),
		moves:  []int{},
	}
	for col := 0; col < ctypes.Width; col += 1 {
		toret.board = append(toret.board, []int{})
		toret.height = append(toret.height, len(s.Columns[col]))
		for row := 0; row < ctypes.Height; row += 1 {
			if row < len(s.Columns[col]) {
				toret.board[col] = append(toret.board[col], int(s.Columns[col][row]))
			} else {
				toret.board[col] = append(toret.board[col], -1)
			}
		}
	}
	return toret
}

func (is *InternalState) GenerateMoves() []int {
	toret := []int{}
	for i, h := range is.height {
		if h < ctypes.Height {
			toret = append(toret, i)
		}
	}
	return toret
}

func (is *InternalState) MakeMove(col int) {
	height := is.height[col]
	is.board[col][height] = is.turn
	is.height[col] = height + 1
	is.turn = 1 - is.turn
	is.moves = append(is.moves, col)
}

func (is *InternalState) UnmakeMove(col int) {
	height := is.height[col]
	is.board[col][height-1] = -1
	is.height[col] = height - 1
	is.turn = 1 - is.turn
	is.moves = is.moves[:len(is.moves)-1]
}

func (is *InternalState) StalemateCheck() bool {
	for _, h := range is.height {
		if h != ctypes.Height {
			return false
		}
	}
	return true
}

func (is *InternalState) directionalWinCheckLastMove(cdelta, rdelta int) int {
	col := is.moves[len(is.moves)-1]
	row := is.height[col] - 1

	rlast := row + (3 * rdelta)
	clast := col + (3 * cdelta)
	if rlast < 0 || clast < 0 {
		return -1
	}
	if rlast >= ctypes.Height || clast >= ctypes.Width {
		return -1
	}

	checkingPiece := is.board[col][row]
	if checkingPiece == -1 {
		return -1
	}
	for i := 1; i < 4; i += 1 {
		rind := row + (i * rdelta)
		cind := col + (i * cdelta)
		if is.board[cind][rind] != checkingPiece {
			return -1
		}
	}
	return checkingPiece
}

func (is *InternalState) directionalWinCheck(cdelta, rdelta int) int {
	for col, h := range is.height {
		for row := 0; row < h; row += 1 {
			rlast := row + (3 * rdelta)
			clast := col + (3 * cdelta)
			if rlast < 0 || clast < 0 {
				continue
			}
			if rlast >= ctypes.Height || clast >= ctypes.Width {
				continue
			}
			match := true
			checkingPiece := is.board[col][row]
			if checkingPiece == -1 {
				continue
			}
			for i := 1; i < 4; i += 1 {
				rind := row + (i * rdelta)
				cind := col + (i * cdelta)
				if is.board[cind][rind] != checkingPiece {
					match = false
					break
				}
			}
			if match {
				return checkingPiece
			}
		}
	}
	return -1
}

func (is *InternalState) VictoryCheck() int {
	v := is.directionalWinCheck(1, 0)
	if v > -1 {
		return v
	}
	v = is.directionalWinCheck(0, 1)
	if v > -1 {
		return v
	}
	v = is.directionalWinCheck(1, 1)
	if v > -1 {
		return v
	}
	return is.directionalWinCheck(-1, 1)
}
