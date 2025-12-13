package priority

type Message struct {
	Priority int
	Data     interface{}
}

type PriorityQueue struct {
	queue []*Message
}

func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		queue: make([]*Message, 0),
	}
}

// Push inserts message keeping the queue ordered by priority (high → low)
func (pq *PriorityQueue) Push(message *Message) {
	if len(pq.queue) == 0 {
		pq.queue = append(pq.queue, message)
		return
	}

	inserted := false
	for i, msg := range pq.queue {
		if message.Priority > msg.Priority {
			// insert at position i
			pq.queue = append(pq.queue[:i], append([]*Message{message}, pq.queue[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.queue = append(pq.queue, message)
	}
}

// Pop removes and returns the highest-priority message
func (pq *PriorityQueue) Pop() *Message {
	if len(pq.queue) == 0 {
		return nil
	}

	msg := pq.queue[0]
	pq.queue = pq.queue[1:]
	return msg
}

// Peek returns the highest-priority message without removing it
func (pq *PriorityQueue) Peek() *Message {
	if len(pq.queue) == 0 {
		return nil
	}
	return pq.queue[0]
}

// Len returns number of items in the queue
func (pq *PriorityQueue) Len() int {
	return len(pq.queue)
}
