package channel

import (
	"sync"
)

const (
	MAX_BUFFER = 1000
)

type Channel struct {
	mux      *sync.Mutex
	internal map[string]chan interface{}
}

func New() *Channel {
	return &Channel{
		mux:      &sync.Mutex{},
		internal: make(map[string]chan interface{}),
	}
}

func (s *Channel) Get(channelName string) chan interface{} {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.internal[channelName] == nil {
		s.internal[channelName] = make(chan interface{}, MAX_BUFFER)
	}
	return s.internal[channelName]
}
