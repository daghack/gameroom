package echo

import (
	"fmt"
)

type EchoGame struct {
	msgs chan []byte
}

func NewEchoGame() *EchoGame {
	msgs := make(chan []byte, 16)
	return &EchoGame{
		msgs: msgs,
	}
}

func (game *EchoGame) Join(playerId string) error {
	fmt.Println(playerId + " has joined the game.")
	return nil
}

func (game *EchoGame) Leave(playerId string) error {
	fmt.Println(playerId + " has left the game.")
	return nil
}

func (game *EchoGame) UpdatesChannel(playerId string) <-chan []byte {
	return game.msgs
}

func (game *EchoGame) MovesChannel(playerId string) chan<- []byte {
	return game.msgs
}
