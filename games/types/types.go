package types

type Move struct {
	playerId string
	data     []bytes
}

type Game interface {
	Join(playerId string) error
	Leave(playerId string) error
	UpdatesChannel(playerId string) <-chan []byte
	MovesChannel(playerId string) chan<- *Move
}
