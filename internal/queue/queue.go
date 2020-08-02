package queue

import (
	"container/list"
	"time"

	"github.com/reposandermets/take-positions/internal/account"
)

var Q = Queue{}

func (q *Queue) Initialize() {
	q.queue = list.New()
	println("Queue initialized")
}

func (q *Queue) Enqueue(p account.Payload) {
	q.queue.PushBack(p)
}

func (q *Queue) Dequeue() {
	for {
		time.Sleep(333 * time.Millisecond)
		e := q.queue.Front()
		if e != nil {
			var payload account.Payload
			payload = e.Value.(account.Payload)
			account.F.HandleQueueItem(payload)
			q.queue.Remove(e)
		}
	}
}
