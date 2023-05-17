package circular

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	cache := New(2)
	assert.Equal(t, int32(0), cache.idx)

	cache.Insert("1")
	assert.Equal(t, int32(1), cache.idx)

	data := cache.Read()
	assert.Equal(t, "1", data[0])

	cache.Insert("2")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Read()
	assert.Equal(t, "1", data[0])
	assert.Equal(t, "2", data[1])

	cache.Insert("3")
	assert.Equal(t, int32(1), cache.idx)

	data = cache.Read()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "2", data[1])

	cache.Insert("4")
	assert.Equal(t, int32(2), cache.idx)

	data = cache.Read()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "4", data[1])
}

func TestSorted(t *testing.T) {
	cache := New(3)
	cache.Insert(1)
	cache.Insert(2)
	cache.Insert(3)
	cache.Insert(4)
	fmt.Println(cache.Sorted()...)
}
