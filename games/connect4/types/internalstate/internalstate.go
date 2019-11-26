package internalstate

import (
	ctypes "websockets/games/connect4/types"
)

type InternalState struct {
	LocScore [][]int
	Board    [][]int
	Height   []int
	turn     int
	Agent    int
	moves    []int
}

func NewInternalState(agentId string, s *ctypes.UpdateGameState) *InternalState {
	loc_score := [][]int{
		[]int{3, 4, 5, 5, 4, 3},
		[]int{4, 6, 8, 8, 6, 4},
		[]int{5, 8, 11, 11, 8, 5},
		[]int{7, 10, 13, 13, 10, 7},
		[]int{5, 8, 11, 11, 8, 5},
		[]int{4, 6, 8, 8, 6, 4},
		[]int{3, 4, 5, 5, 4, 3},
	}
	toret := &InternalState{
		LocScore: loc_score,
		Board:    [][]int{},
		Height:   []int{},
		turn:     int(s.CurrentTurn),
		Agent:    int(s.Players[agentId]),
		moves:    []int{},
	}
	for col := 0; col < ctypes.Width; col += 1 {
		toret.Board = append(toret.Board, []int{})
		toret.Height = append(toret.Height, len(s.Columns[col]))
		for row := 0; row < ctypes.Height; row += 1 {
			if row < len(s.Columns[col]) {
				toret.Board[col] = append(toret.Board[col], int(s.Columns[col][row]))
			} else {
				toret.Board[col] = append(toret.Board[col], -1)
			}
		}
	}
	return toret
}

func (is *InternalState) ToString() string {
	var toret [ctypes.Width * ctypes.Height]byte
	for col_i, col := range is.Board {
		for row_i, piece := range col {
			index := col_i*ctypes.Height + row_i
			piece_p := uint8(piece + 1)
			toret[index] = piece_p + 48
		}
	}
	return string(toret[:])
}

func (is *InternalState) GenerateMoves() []int {
	toret := []int{}
	for i, h := range is.Height {
		if h < ctypes.Height {
			toret = append(toret, i)
		}
	}
	return toret
}

func (is *InternalState) MakeMove(col int) {
	Height := is.Height[col]
	is.Board[col][Height] = is.turn
	is.Height[col] = Height + 1
	is.turn = 1 - is.turn
	is.moves = append(is.moves, col)
}

func (is *InternalState) UnmakeMove(col int) {
	Height := is.Height[col]
	is.Board[col][Height-1] = -1
	is.Height[col] = Height - 1
	is.turn = 1 - is.turn
	is.moves = is.moves[:len(is.moves)-1]
}

func (is *InternalState) StalemateCheck() bool {
	for _, h := range is.Height {
		if h != ctypes.Height {
			return false
		}
	}
	return true
}

func (is *InternalState) directionalWinCheckLastMove(cdelta, rdelta int) int {
	col := is.moves[len(is.moves)-1]
	row := is.Height[col] - 1

	rlast := row + (3 * rdelta)
	clast := col + (3 * cdelta)
	if rlast < 0 || clast < 0 {
		return -1
	}
	if rlast >= ctypes.Height || clast >= ctypes.Width {
		return -1
	}

	checkingPiece := is.Board[col][row]
	if checkingPiece == -1 {
		return -1
	}
	for i := 1; i < 4; i += 1 {
		rind := row + (i * rdelta)
		cind := col + (i * cdelta)
		if is.Board[cind][rind] != checkingPiece {
			return -1
		}
	}
	return checkingPiece
}

func (is *InternalState) directionalWinCheck(cdelta, rdelta int) int {
	for col, h := range is.Height {
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
			checkingPiece := is.Board[col][row]
			if checkingPiece == -1 {
				continue
			}
			for i := 1; i < 4; i += 1 {
				rind := row + (i * rdelta)
				cind := col + (i * cdelta)
				if is.Board[cind][rind] != checkingPiece {
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
