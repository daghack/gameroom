package main

import (
	"websockets/ai"
	"websockets/ai/minmax/connect4ai"
)

func main() {
	a := &connect4ai.Agent{
		AgentId: "minmax",
	}
	agent, err := ai.NewAgent(a, "ws://localhost:8080/game?userId=minmax")
	if err != nil {
		panic(err)
	}
	agent.Run()
}
