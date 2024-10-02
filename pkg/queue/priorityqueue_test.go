package queue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue(t *testing.T) {
	t.Run("Send/Recv", func(t *testing.T) {
		messages := []*Message[string]{
			&Message[string]{
				Priority: 0,
				Value:    "E",
			},
			&Message[string]{
				Priority: 5,
				Value:    "D",
			},
			&Message[string]{
				Priority: 7,
				Value:    "B",
			},
			&Message[string]{
				Priority: 5,
				Value:    "C",
			},
			&Message[string]{
				Priority: 9,
				Value:    "A",
			},
		}

		pq := NewPriorityQueue[string](len(messages))

		var wg sync.WaitGroup
		var actual []string

		for _, msg := range messages {
			wg.Add(1)
			pq.Send(msg)
		}

		go pq.Recv(func(msg *Message[string]) {
			actual = append(actual, msg.Value)
			wg.Done()
		})

		wg.Wait()
		expected := []string{"A", "B", "C", "D", "E"}
		assert.Equal(t, expected, actual)
	})
}
