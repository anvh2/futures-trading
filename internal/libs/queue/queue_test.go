package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := New()

	type sample struct {
		id   int
		name string
	}

	q.Push(context.Background(), "foo", &sample{id: 1, name: "foo"})
	data, err := q.Consume(context.Background(), "bar", "group1")
	assert.Nil(t, data)
	assert.Equal(t, ErrNoMessageAvailable, err)

	data, err = q.Consume(context.Background(), "foo", "group1")
	assert.Nil(t, err)
	assert.Equal(t, &sample{id: 1, name: "foo"}, data.Data)

	offset := data.Offset

	data, err = q.Consume(context.Background(), "foo", "group1")
	assert.Nil(t, data)
	assert.Equal(t, ErrMustCommitBeforeConsuming, err)

	q.Commit(context.Background(), "foo", "group1", offset)

	data, err = q.Consume(context.Background(), "foo", "group1")
	assert.Nil(t, data)
	assert.Equal(t, ErrNoMessageAvailable, err)
}
