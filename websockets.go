package main

import (
	"net/http"
	"websockets/gameroom"
	"websockets/games/connect4"
)

var game *gameroom.GameRoom

func gameConnect(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if userid, ok := query["userId"]; ok {
		game.ConnectToGame(userid[0], w, r)
	} else {
		http.Error(w, "No User Id", http.StatusForbidden)
	}
}

func main() {
	game, _ = gameroom.NewGameRoom(connect4.NewConnect4())
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/game", gameConnect)
	http.ListenAndServe(":8080", nil)
}
