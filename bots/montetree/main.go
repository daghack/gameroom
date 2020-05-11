package main

import (
	"os"
	"websockets/ai"
	"websockets/ai/montecarlotree/connect4ai"
)

func main() {
	id := "montetree"
	if len(os.Args) > 1 {
		id = os.Args[1]
	}
	a := connect4ai.NewAgent(id)
	agent, err := ai.NewAgent(a, "ws://localhost:8080/game?userId="+id)
	if err != nil {
		panic(err)
	}
	agent.Run()
}
