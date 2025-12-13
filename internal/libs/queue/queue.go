package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/lease"
)

var (
	ErrInvalidMessage            = errors.New("invalid message")
	ErrNoMessageAvailable        = errors.New("no message available")
	ErrExpireNegative            = errors.New("expire time must be positive")
	ErrMustCommitBeforeConsuming = errors.New("must commit before consuming")
)

var _ IQueue = &Queue{}

type IQueue interface {
	Push(ctx context.Context, topicName string, data interface{}, opts ...OptionFunc) error
	Consume(ctx context.Context, topicName, groupID string) (*Message, error)
	Commit(ctx context.Context, topicName, groupID string, offset int64)
	Close()
}

const (
	defaultRetention = time.Hour
	cleanupInterval  = 30 * time.Second
)

type Message struct {
	Topic   string
	GroupID string
	Offset  int64
	Data    interface{}

	expire time.Time
	commit func(ctx context.Context, topicName, groupID string, offset int64)
}

func (m *Message) Commit(ctx context.Context) error {
	if m == nil || m.commit == nil {
		return ErrInvalidMessage
	}

	m.commit(ctx, m.Topic, m.GroupID, m.Offset)
	m.commit = nil
	return nil
}

type ConsumerGroup struct {
	leases map[string]*lease.Lease // topic -> lease

	GroupID string
	Offsets map[string]int64 // topic -> offset
}

type Topic struct {
	name   string
	length int64
	table  map[int64]*Message
	lock   sync.Mutex
}

type Queue struct {
	topics    map[string]*Topic
	groups    map[string]*ConsumerGroup
	lock      sync.Mutex
	quit      chan struct{}
	retention time.Duration
}

func New(opts ...OptionFunc) *Queue {
	option := configure(opts...)

	q := &Queue{
		topics:    make(map[string]*Topic),
		groups:    make(map[string]*ConsumerGroup),
		quit:      make(chan struct{}),
		retention: option.retention,
	}

	go q.cleanup()
	return q
}

func (q *Queue) getOrCreateTopic(name string) *Topic {
	q.lock.Lock()
	defer q.lock.Unlock()

	if topic, ok := q.topics[name]; ok {
		return topic
	}

	topic := &Topic{
		name:  name,
		table: make(map[int64]*Message),
	}
	q.topics[name] = topic
	return topic
}

func (q *Queue) getOrCreateGroup(groupID string) *ConsumerGroup {
	q.lock.Lock()
	defer q.lock.Unlock()

	if g, ok := q.groups[groupID]; ok {
		return g
	}

	g := &ConsumerGroup{
		leases:  make(map[string]*lease.Lease),
		GroupID: groupID,
		Offsets: make(map[string]int64),
	}
	q.groups[groupID] = g
	return g
}

func (q *Queue) Push(ctx context.Context, topicName string, data interface{}, opts ...OptionFunc) error {
	option := configure(opts...)
	expire := option.expire
	if expire == 0 {
		expire = q.retention
	}

	topic := q.getOrCreateTopic(topicName)

	topic.lock.Lock()
	defer topic.lock.Unlock()

	topic.length++
	msg := &Message{
		Offset: topic.length,
		Data:   data,
		expire: time.Now().Add(expire),
	}

	topic.table[msg.Offset] = msg
	return nil
}

func (q *Queue) Consume(ctx context.Context, topicName, groupID string) (*Message, error) {
	topic := q.getOrCreateTopic(topicName)
	group := q.getOrCreateGroup(groupID)

	ls, ok := group.leases[topicName]
	if !ok {
		ls = lease.New()
		group.leases[topicName] = ls
	}

	acquired := ls.Try()
	if !acquired {
		return nil, ErrMustCommitBeforeConsuming
	}

	topic.lock.Lock()
	defer topic.lock.Unlock()

	offset := group.Offsets[topicName] + 1

	for offset <= topic.length {
		msg, ok := topic.table[offset]
		if !ok {
			offset++
			group.Offsets[topicName] = offset - 1
			continue
		}

		if msg.expire.Before(time.Now()) {
			delete(topic.table, offset)
			offset++
			group.Offsets[topicName] = offset - 1
			continue
		}

		msg.GroupID = groupID
		msg.Topic = topicName
		msg.commit = q.Commit

		return msg, nil
	}

	ls.Release()
	return nil, ErrNoMessageAvailable
}

func (q *Queue) Commit(ctx context.Context, topicName, groupID string, offset int64) {
	group := q.getOrCreateGroup(groupID)
	group.Offsets[topicName] = offset
	if lease, ok := group.leases[topicName]; ok {
		lease.Release()
	}
}

func (q *Queue) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			q.cleanupExpired()
		case <-q.quit:
			return
		}
	}
}

func (q *Queue) cleanupExpired() {
	q.lock.Lock()
	defer q.lock.Unlock()

	now := time.Now()

	for _, topic := range q.topics {
		topic.lock.Lock()
		for offset, msg := range topic.table {
			if msg.expire.Before(now) {
				delete(topic.table, offset)
			}
		}
		topic.lock.Unlock()
	}
}

func (q *Queue) Close() {
	close(q.quit)
}
