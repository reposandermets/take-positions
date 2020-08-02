package core

import (
	"github.com/reposandermets/take-positions/internal/account"
	"github.com/reposandermets/take-positions/internal/queue"
)

func Run() {
	queue.Q.Initialize()
	go queue.Q.Dequeue()
	account.F.Initialize()
}
