package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	q := New()

	type sample struct {
		id   int
		name string
	}

	q.Push(&sample{id: 1, name: "foo"}, time.Second)
	data, err := q.Peak("bar")
	fmt.Println(data.Data, err)
}
