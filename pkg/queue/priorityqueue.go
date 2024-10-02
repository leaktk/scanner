package queue

import (
	"container/heap"
	"sync"
)

// PriorityQueue is like a channel but with dynamic buffering and returns items
// with the highest priority first
type PriorityQueue[T any] struct {
	heap      *MessageHeap[T]
	heapMutex sync.Mutex
	out       chan *Message[T]
	msgCond   *sync.Cond
}

// NewPriorityQueue returns a PriorityQueue instance that is ready to send to
func NewPriorityQueue[T any](queueSize int) *PriorityQueue[T] {
	pq := &PriorityQueue[T]{
		heap:    NewMessageHeap[T](queueSize),
		out:     make(chan *Message[T]),
		msgCond: sync.NewCond(&sync.Mutex{}),
	}

	// Init the heap
	heap.Init(pq.heap)

	// Set up message forwarding
	go func() {
		for {
			pq.heapMutex.Lock()
			count := pq.heap.Len()
			pq.heapMutex.Unlock()

			if count == 0 {
				pq.waitForMessage()
			}

			// Since out can block, only lock for popping the message
			pq.heapMutex.Lock()
			msg := heap.Pop(pq.heap).(*Message[T])
			pq.heapMutex.Unlock()

			pq.out <- msg
		}
	}()

	return pq
}

// Send puts items on the queue
func (pq *PriorityQueue[T]) Send(msg *Message[T]) {
	pq.heapMutex.Lock()
	heap.Push(pq.heap, msg)
	pq.heapMutex.Unlock()
	pq.signalMessageRecieved()
}

// Recv takes a function that can recieve messages sent to the queue
func (pq *PriorityQueue[T]) Recv(fn func(*Message[T])) {
	for msg := range pq.out {
		fn(msg)
	}
}

func (pq *PriorityQueue[T]) waitForMessage() {
	pq.msgCond.L.Lock()
	pq.msgCond.Wait()
	pq.msgCond.L.Unlock()
}

func (pq *PriorityQueue[T]) signalMessageRecieved() {
	pq.msgCond.L.Lock()
	pq.msgCond.Signal()
	pq.msgCond.L.Unlock()
}
