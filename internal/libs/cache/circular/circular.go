package circular

import (
	"sync"
)

type Cache struct {
	head  int32         // Write position
	tail  int32         // Read position
	count int32         // Current elements
	size  int32         // Buffer capacity
	data  []interface{} // Pre-allocated slice (faster than map)
	mutex *sync.RWMutex
}

func New(size int32) *Cache {
	return &Cache{
		head:  0,
		tail:  0,
		count: 0,
		size:  size,
		data:  make([]interface{}, size), // Pre-allocate for better performance
		mutex: &sync.RWMutex{},
	}
}

func (l *Cache) Insert(data interface{}) int32 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	pos := l.head
	l.data[pos] = data

	l.head = (l.head + 1) % l.size
	if l.count < l.size {
		l.count++
	} else {
		l.tail = (l.tail + 1) % l.size // Overwrite oldest when full
	}

	return pos
}

func (l *Cache) Update(idx int32, data interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.data[idx] = data
}

func (l *Cache) Read() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]interface{}, l.count)
	for i := int32(0); i < l.count; i++ {
		data[i] = l.data[i]
	}

	return data
}

func (l *Cache) Tail() (interface{}, int32) {
	if l == nil || l.count == 0 {
		return nil, -1
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	var idx int32

	if l.head == 0 {
		idx = l.size - 1
	} else {
		idx = l.head - 1
	}

	return l.data[idx], idx
}

func (l *Cache) Sorted() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.count == 0 {
		return []interface{}{}
	}

	data := make([]interface{}, l.count)
	idx := 0

	// Start from tail and read count elements
	for i := int32(0); i < l.count; i++ {
		pos := (l.tail + i) % l.size
		data[idx] = l.data[pos]
		idx++
	}

	return data
}
