package types

type Move struct {
	PlayerId string
	Data     []byte
}

type Game interface {
	Join(playerId string) error
	Leave(playerId string) error
	UpdatesChannel(playerId string) (<-chan []byte, error)
	MovesChannel(playerId string) (chan<- *Move, error)
	Close()
}
