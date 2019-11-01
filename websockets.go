package main

import (
	"net/http"
	"websockets/gameroom"
	"websockets/games/connect4"
)

var game *gameroom.GameRoom

func gameConnect(w http.ResponseWriter, r *http.Request) {
	game.ConnectToGame("super id", w, r)
}

func main() {
	game, _ = gameroom.NewGameRoom(connect4.NewConnect4())
	http.HandleFunc("/game", gameConnect)
	http.ListenAndServe("localhost:8080", nil)
}
