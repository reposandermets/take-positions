package core

import (
	"container/list"
	"time"
)

var Q = Queue{}

func (q *Queue) Initialize() {
	q.queue = list.New()
	println("Queue initialized")
}

func (q *Queue) Enqueue(p Payload) {
	q.queue.PushBack(p)
}

func (q *Queue) Dequeue() {
	for {
		time.Sleep(333 * time.Millisecond)
		e := q.queue.Front()
		if e != nil {
			var payload Payload
			payload = e.Value.(Payload)
			F.HandleQueueItem(payload)
			q.queue.Remove(e)
		}
	}
}

func Run() {
	Q.Initialize()
}
