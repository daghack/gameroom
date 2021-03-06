package main

import (
	"os"
	"websockets/ai"
	"websockets/ai/minmax/connect4ai"
)

func main() {
	id := "minmax"
	if len(os.Args) > 1 {
		id = os.Args[1]
	}
	a := &connect4ai.Agent{
		AgentId: id,
	}
	agent, err := ai.NewAgent(a, "ws://localhost:8080/game?userId="+id)
	if err != nil {
		panic(err)
	}
	agent.Run()
}
