package connect4

import (
	"fmt"
	"games/types"
)

const (
	Red = iota
	Black
)

type color int

type Connect4 struct {
	players map[string]color
}

func NewConnect4() *Connect4 {
	return &Connect4{
		players: map[string]color{},
	}
}

func (connect *Connect4) Join(playerId string) error {
	if _, ok := connect.players[playerId]; ok {
		return nil
	}
	if len(connect.players) >= 2 {
		return fmt.Errorf("Game Filled")
	}
	connect.players[playerId] = color(len(connect.players))
	return nil
}

func (connect *Connect4) Leave(playerId string) error {
	delete(connect.players, playerId)
}

func (connect *Connect4) UpdatesChannel(playerId string) <-chan []byte {
	return nil
}

func (connect *Connect4) MovesChannel(playerId string) chan<- *types.Move {
	return nil
}
