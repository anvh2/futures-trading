package circular

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSet(t *testing.T) {
	cache := New(2)
	assert.Equal(t, int32(0), cache.head)

	cache.Insert("1")
	assert.Equal(t, int32(1), cache.head)

	data := cache.Read()
	assert.Equal(t, "1", data[0])

	cache.Insert("2")
	assert.Equal(t, int32(0), cache.head) // Wrapped around

	data = cache.Read()
	assert.Equal(t, "1", data[0])
	assert.Equal(t, "2", data[1])

	cache.Insert("3") // Overwrites oldest
	assert.Equal(t, int32(1), cache.head)

	data = cache.Read()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "2", data[1])

	cache.Insert("4")
	assert.Equal(t, int32(0), cache.head)

	data = cache.Read()
	assert.Equal(t, "3", data[0])
	assert.Equal(t, "4", data[1])
}

func TestSorted(t *testing.T) {
	cache := New(3)
	cache.Insert(1)
	cache.Insert(2)
	cache.Insert(3)
	cache.Insert(4) // Overwrites oldest (1)

	sorted := cache.Sorted()
	// Should return in chronological order: [2, 3, 4]
	assert.Equal(t, 3, len(sorted))
	assert.Equal(t, 2, sorted[0])
	assert.Equal(t, 3, sorted[1])
	assert.Equal(t, 4, sorted[2])

	fmt.Printf("Sorted order: %v\n", sorted)
}
