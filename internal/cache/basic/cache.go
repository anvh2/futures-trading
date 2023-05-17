package basic

import (
	"sync"
)

type Cache struct {
	lock     *sync.RWMutex
	list     []interface{}
	internal map[string]interface{}
}

func NewCache() *Cache {
	return &Cache{
		lock:     &sync.RWMutex{},
		list:     []interface{}{},
		internal: make(map[string]interface{}),
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.internal[key] = value
	c.list = append(c.list, value)
}

func (c *Cache) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.internal[key]
}

func (c *Cache) Exs(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.internal[key]
	return ok
}

// SetEX sets and check exists
func (c *Cache) SetEX(key string, value interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if old, ok := c.internal[key]; ok {
		c.internal[key] = value
		return old, true
	}

	c.internal[key] = value
	return value, false
}
