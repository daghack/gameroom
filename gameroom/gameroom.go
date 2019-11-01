package gameroom

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"net/http"
	"websocket/games/types"
)

type GameRoom struct {
	Id                string
	game              types.Game
	playerConnections map[string]chan bool
}

func NewGameRoom(game types.Game) (*GameRoom, error) {
	id := uuid.NewV4()
	return &GameRoom{
		Id:                id.String(),
		game:              game,
		playerConnections: map[string]chan bool{},
	}, nil
}

func (gr *GameRoom) ConnectToGame(playerId string, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Actually, return some sort of error status.
		panic(err)
	}

	err = gr.game.Join(playerId)
	if err != nil {
		panic(err)
	}
	defer gr.game.Leave(playerId)

	gr.runPlayerSession(playerId, c)
}

func (gr *GameRoom) runPlayerSession(playerId string, conn *websocket.Conn) {
	_, ok := gr.playerConnections[playerId]
	if ok {
		// Already opened in another browser, probably. Figure out how to handle this?
		return
	}
	ch := make(chan bool, 1)
	gr.playerConnections[playerId] = ch
	go gr.forwardGameUpdates(playerId, conn)
	gr.forwardGameMoves(playerId, conn)
	ch <- true
}

func (gr *GameRoom) forwardGameMoves(playerId string, conn *websocket.Conn) {
	moveChan := gr.game.MovesChannel(playerId)
	for {
		mtype, msg, err := conn.ReadMessage()
		fmt.Println(mtype, string(msg), err)
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				return
			default:
				panic(err)
			}
		}
		switch mtype {
		case websocket.TextMessage:
			move := &types.Move{
				playerId: playerId,
				data:     msg,
			}
			moveChan <- move
		default:
			continue
		}
	}
}

func (gr *GameRoom) forwardGameUpdates(playerId string, conn *websocket.Conn) {
	updates := gr.game.UpdatesChannel(playerId)
	for {
		select {
		case <-gr.playerConnections[playerId]:
			// We've received word that the connection is closed
			return
		case update := <-updates:
			err := conn.WriteMessage(websocket.TextMessage, update)
			if err != nil {
				// Can't write, connection is probably closed
				return
			}
		}
	}
}
