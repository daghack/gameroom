package main

import (
	"net/http"
	"websockets/ai"
	"websockets/ai/minmax/connect4ai"
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
	http.HandleFunc("/game", gameConnect)
	go http.ListenAndServe("localhost:8080", nil)
	a := &connect4ai.Agent{
		AgentId: "Bot",
	}
	agent, err := ai.NewAgent(a, "ws://localhost:8080/game?userId=Bot")
	if err != nil {
		panic(err)
	}
	agent.Run()
}
