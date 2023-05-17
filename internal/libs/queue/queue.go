package queue

import (
	"errors"
	"sync"
	"time"
)

const (
	defaultRetention time.Duration = time.Hour
)

type Message struct {
	expire time.Time
	Offset int64
	Data   interface{}
}

type Consumer struct {
	ConsumerId    string
	currentOffset int64
}

type Queue struct {
	lock      *sync.Mutex
	length    int64
	table     map[int64]*Message
	consumers map[string]*Consumer
	quit      chan struct{}
}

func New() *Queue {
	queue := &Queue{
		lock:      &sync.Mutex{},
		length:    0,
		table:     make(map[int64]*Message),
		consumers: make(map[string]*Consumer),
		quit:      make(chan struct{}),
	}

	// ensure there are no memory leaks
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer recover()

		for {
			select {
			case <-ticker.C:
				for offset, msg := range queue.table {
					if msg.expire.Before(time.Now()) {
						queue.remove(offset)
					}
				}

			case <-queue.quit:
				return
			}
		}
	}()

	return queue
}

func (q *Queue) remove(offset int64) {
	q.lock.Lock()
	defer q.lock.Unlock()

	delete(q.table, offset)
}

func (q *Queue) Register(consumerId string) *Consumer {
	q.lock.Lock()
	defer q.lock.Unlock()

	consumer := &Consumer{
		ConsumerId:    consumerId,
		currentOffset: 0,
	}
	q.consumers[consumerId] = consumer

	return consumer
}

func (q *Queue) Push(data interface{}, expire time.Duration) error {
	if expire.Milliseconds() < 0 {
		return errors.New("expire time negative")
	}

	if expire.Milliseconds() == 0 {
		expire = defaultRetention
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.length++

	msg := &Message{
		expire: time.Now().Add(expire),
		Offset: q.length,
		Data:   data,
	}

	q.table[q.length] = msg

	return nil
}

func (q *Queue) Peak(consumerId string) (*Message, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	consumer, ok := q.consumers[consumerId]
	if !ok {
		consumer = q.Register(consumerId)
	}

	for consumer.currentOffset <= q.length {
		msg, ok := q.table[consumer.currentOffset]
		if !ok {
			consumer.currentOffset++
			continue
		}

		if msg.expire.Before(time.Now()) {
			delete(q.table, consumer.currentOffset)
			consumer.currentOffset++
			continue
		}

		return msg, nil
	}

	return nil, errors.New("notfound")
}

func (q *Queue) Close() {
	close(q.quit)
}
