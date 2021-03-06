package connect4

import (
	"encoding/json"
	"fmt"
	ctypes "websockets/games/connect4/types"
	"websockets/games/types"
)

type Connect4 struct {
	state       ctypes.GameState
	players     map[string]*ctypes.PlayerInfo
	moveChannel chan *types.Move
}

func NewConnect4() *Connect4 {
	toret := &Connect4{
		state:       ctypes.NewGameState(),
		players:     map[string]*ctypes.PlayerInfo{},
		moveChannel: make(chan *types.Move, 16),
	}
	go toret.gameLoop()
	return toret
}

func (connect *Connect4) playerByPieceColor(piece ctypes.Color) string {
	for id, playerInfo := range connect.players {
		if playerInfo.PlayerColor == piece {
			return id
		}
	}
	return ""
}

func (connect *Connect4) requestRematch(player *ctypes.PlayerInfo) {
	player.RematchAttempt = true
	rematch := true
	for _, player := range connect.players {
		rematch = rematch && player.RematchAttempt
	}
	if rematch {
		connect.state = ctypes.NewGameState()
		for _, player := range connect.players {
			player.RematchAttempt = false
		}
	}
}

func (connect *Connect4) gameLoop() {
	for move := range connect.moveChannel {
		info, ok := connect.players[move.PlayerId]
		if !ok {
			continue
		}

		m := &ctypes.MoveData{
			Col: -1,
		}
		err := json.Unmarshal(move.Data, m)
		if err != nil || (m.Col < 0 && !m.Rematch) {
			continue
		}

		if m.Rematch {
			connect.requestRematch(info)
		} else {
			err = connect.makeMove(info.PlayerColor, m.Col)
			if err != nil {
				continue
			}
		}
		connect.sendUpdates()
	}
}

func (connect *Connect4) sendUpdates() {
	update := connect.marshalState()
	for _, info := range connect.players {
		info.UpdateChan <- update
	}
}

func (connect *Connect4) makeMove(piece ctypes.Color, col int) error {
	if piece != connect.state.CurrentTurn {
		return fmt.Errorf("Not the correct turn.")
	}
	if col >= ctypes.Width || col < 0 {
		return fmt.Errorf("Not a legitimate move")
	}
	if len(connect.state.Columns[col]) >= ctypes.Height {
		return fmt.Errorf("Column Full")
	}
	if connect.state.GameOver {
		return nil
	}
	lastTurn := connect.state.CurrentTurn
	connect.state.CurrentTurn = ctypes.Black - connect.state.CurrentTurn
	connect.state.Columns[col] = append(connect.state.Columns[col], piece)
	winCheck := connect.winCheck(col)
	if winCheck != nil {
		player := connect.playerByPieceColor(lastTurn)
		fmt.Println("WINNER:", player)
		connect.state.GameOver = true
		connect.state.WinningPositions = winCheck
	}
	return nil
}

func (connect *Connect4) marshalState() []byte {
	state := &ctypes.UpdateGameState{
		GameState: connect.state,
		Players:   map[string]ctypes.Color{},
	}
	for player, info := range connect.players {
		state.Players[player] = info.PlayerColor
	}
	stateJson, _ := json.Marshal(state)
	return stateJson
}

func (connect *Connect4) Close() {
	close(connect.moveChannel)
}

func (connect *Connect4) Join(playerId string) error {
	if _, ok := connect.players[playerId]; ok {
		fmt.Println("PLAYER " + playerId + " ALREADY IN GAME")
	} else if len(connect.players) >= 2 {
		fmt.Println("PLAYER" + playerId + " CAN'T JOIN GAME, GAME FILLED.")
		return fmt.Errorf("Game Filled")
	} else {
		fmt.Println("PLAYER " + playerId + " JOINED GAME")
		connect.players[playerId] = &ctypes.PlayerInfo{
			PlayerColor: ctypes.Color(len(connect.players)),
			UpdateChan:  make(chan []byte, 16),
		}
	}
	update := connect.marshalState()
	connect.players[playerId].UpdateChan <- update
	return nil
}

func (connect *Connect4) Leave(playerId string) error {
	fmt.Println("PLAYER " + playerId + " LEFT GAME")
	delete(connect.players, playerId)
	return nil
}

func (connect *Connect4) UpdatesChannel(playerId string) (<-chan []byte, error) {
	if info, ok := connect.players[playerId]; ok {
		return info.UpdateChan, nil
	}
	return nil, fmt.Errorf("No player in game with id %s", playerId)
}

func (connect *Connect4) MovesChannel(playerId string) (chan<- *types.Move, error) {
	return connect.moveChannel, nil
}

func (connect *Connect4) winCheck(lastPlayed int) []ctypes.Position {
	if lastPlayed < 0 || lastPlayed >= ctypes.Width {
		return nil
	}
	winCheck := connect.verticalCheck(lastPlayed)
	if winCheck != nil {
		return winCheck
	}
	winCheck = connect.horizontalCheck(lastPlayed)
	if winCheck != nil {
		return winCheck
	}
	return connect.diagonalCheck(lastPlayed)
}

func (connect *Connect4) verticalCheck(lastPlayed int) []ctypes.Position {
	colLen := len(connect.state.Columns[lastPlayed])
	if colLen < 4 {
		return nil
	}
	lastPiece := connect.state.Columns[lastPlayed][colLen-1]
	top := connect.state.Columns[lastPlayed][colLen-4:]
	win := true
	for _, piece := range top {
		if piece != lastPiece {
			win = false
		}
	}
	if win {
		toret := []ctypes.Position{}
		for i := colLen - 4; i < colLen; i += 1 {
			toret = append(toret, ctypes.Position{
				Col: lastPlayed,
				Row: i,
			})
		}
		return toret
	}
	return nil
}

func (connect *Connect4) horizontalCheck(lastPlayed int) []ctypes.Position {
	colLen := len(connect.state.Columns[lastPlayed])
	lastPiece := connect.state.Columns[lastPlayed][colLen-1]
	minC := lastPlayed - 3
	if minC < 0 {
		minC = 0
	}
	maxC := lastPlayed + 3
	if maxC >= ctypes.Width {
		maxC = ctypes.Width - 1
	}
	for col := minC; col <= lastPlayed; col += 1 {
		if col+3 > maxC {
			return nil
		}
		win := true
		for colC := col; colC < col+4; colC += 1 {
			if len(connect.state.Columns[colC]) < colLen || connect.state.Columns[colC][colLen-1] != lastPiece {
				win = false
				break
			}
		}
		if win {
			toret := []ctypes.Position{}
			for colC := col; colC < col+4; colC += 1 {
				toret = append(toret, ctypes.Position{
					Row: colLen - 1,
					Col: colC,
				})
			}
			return toret
		}
	}
	return nil
}

func (connect *Connect4) diagonalCheck(pieceCol int) []ctypes.Position {
	pieceRow := len(connect.state.Columns[pieceCol]) - 1
	lastPiece := connect.state.Columns[pieceCol][pieceRow]
	colLeft := pieceCol - 3
	rowBot := pieceRow - 3
	rowTop := pieceRow + 3
	for i := 0; i < 7; i += 1 {
		row_i := rowBot + i
		col_i := colLeft + i
		if row_i < 0 || col_i < 0 || row_i >= ctypes.Height || col_i >= ctypes.Width {
			continue
		}
		win := true
		for check := 0; check < 4; check += 1 {
			if row_i+check >= ctypes.Height || col_i+check >= ctypes.Width {
				win = false
				break
			}
			colHeight := len(connect.state.Columns[col_i+check])
			if row_i+check >= colHeight {
				win = false
				break
			}
			piece := connect.state.Columns[col_i+check][row_i+check]
			if piece != lastPiece {
				win = false
				break
			}
		}
		if win {
			toret := []ctypes.Position{}
			for check := 0; check < 4; check += 1 {
				toret = append(toret, ctypes.Position{
					Row: row_i + check,
					Col: col_i + check,
				})
			}
			return toret
		}
	}
	for i := 0; i < 7; i += 1 {
		row_i := rowTop - i
		col_i := colLeft + i
		if row_i < 0 || col_i < 0 || row_i >= ctypes.Height || col_i >= ctypes.Width {
			continue
		}
		win := true
		for check := 0; check < 4; check += 1 {
			if row_i-check < 0 || col_i+check >= ctypes.Width {
				win = false
				break
			}
			colHeight := len(connect.state.Columns[col_i+check])
			if row_i-check >= colHeight {
				win = false
				break
			}
			piece := connect.state.Columns[col_i+check][row_i-check]
			if piece != lastPiece {
				win = false
				break
			}
		}
		if win {
			toret := []ctypes.Position{}
			for check := 0; check < 4; check += 1 {
				toret = append(toret, ctypes.Position{
					Row: row_i - check,
					Col: col_i + check,
				})
			}
			return toret
		}
	}
	return nil
}
