package circular

import (
	"sync"
)

// should be use first and last instead of idx and len
type Cache struct {
	idx   int32
	len   int32
	size  int32
	data  map[int32]interface{}
	mutex *sync.RWMutex
}

func New(size int32) *Cache {
	return &Cache{
		idx:   0,
		len:   0,
		size:  size,
		data:  make(map[int32]interface{}, size),
		mutex: &sync.RWMutex{},
	}
}

func (l *Cache) Insert(data interface{}) int32 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.idx >= l.size {
		l.idx -= l.size
	}

	l.data[l.idx] = data
	l.idx++

	if l.len < l.size {
		l.len++
	}

	return l.idx - 1
}

func (l *Cache) Update(idx int32, data interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.data[idx] = data
}

func (l *Cache) Read() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]interface{}, l.len)
	for i := int32(0); i < l.len; i++ {
		data[i] = l.data[i]
	}

	return data
}

func (l *Cache) Tail() (interface{}, int32) {
	if l == nil {
		return nil, -1
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	var idx int32

	if l.idx == 0 {
		idx = l.size - 1
	} else {
		idx = l.idx - 1
	}

	return l.data[idx], idx
}

func (l *Cache) Sorted() []interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	idx := 0
	data := make([]interface{}, l.len)

	for i := l.idx; i < l.len; i++ {
		data[idx] = l.data[i]
		idx++
	}

	for i := int32(0); i < l.idx; i++ {
		data[idx] = l.data[i]
		idx++
	}

	return data
}
