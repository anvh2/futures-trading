package lease

import (
	"sync"
	"time"
)

const leaseDuration = 30 * time.Second

type Lease struct {
	mu        sync.Mutex
	locked    bool
	expiresAt time.Time
}

func New() *Lease {
	return &Lease{
		mu:        sync.Mutex{},
		locked:    false,
		expiresAt: time.Time{},
	}
}

func (l *Lease) Try() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if l.locked && now.Before(l.expiresAt) {
		return false
	}

	l.locked = true
	l.expiresAt = now.Add(leaseDuration)
	return true
}

func (l *Lease) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.locked = false
}

func (l *Lease) Expired() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return time.Now().After(l.expiresAt)
}
