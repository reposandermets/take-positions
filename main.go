package main

import (
	"github.com/reposandermets/take-positions/api"
	core "github.com/reposandermets/take-positions/internal"
)

func main() {
	core.Run()
	go core.Q.Dequeue()
	core.RunFlow()
	api.Run()
}
