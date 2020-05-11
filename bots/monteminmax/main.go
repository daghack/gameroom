package main

import (
	"websockets/ai"
	"websockets/ai/monteminmax/connect4ai"
)

func main() {
	a := &connect4ai.Agent{
		AgentId: "monteminmax",
	}
	agent, err := ai.NewAgent(a, "ws://localhost:8080/game?userId=monteminmax")
	if err != nil {
		panic(err)
	}
	agent.Run()
}
