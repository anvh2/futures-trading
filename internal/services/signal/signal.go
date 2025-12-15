package signal

import (
	"sync"

	"github.com/anvh2/futures-trading/internal/libs/heap"
	"github.com/anvh2/futures-trading/internal/models"
)

var _ Service = (*SignalService)(nil)

// SignalService is the core orchestrator that manages multiple interval-isolated
// signal queues. Each interval (e.g., "1m", "5m", "1h") has its own priority queue
// to prevent cross-pollution between different timeframes.
type SignalService struct {
	mu        sync.RWMutex            // protects concurrent access to queues
	queues    map[string]*heap.LPHeap // interval -> queue mapping
	maxPerInt int                     // maximum signals per interval
}

// NewSignalService creates a new signal service with the specified maximum
// number of signals per interval. Each interval gets its own isolated queue.
func NewService(maxPerInterval int) *SignalService {
	return &SignalService{
		queues:    make(map[string]*heap.LPHeap),
		maxPerInt: maxPerInterval,
	}
}

// Ingest adds a signal to the appropriate interval queue.
// If no queue exists for the signal's interval, a new one is created.
// This method is thread-safe and can be called concurrently.
func (s *SignalService) Ingest(signal *models.Signal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get or create queue for this interval
	q, ok := s.queues[signal.Interval]
	if !ok {
		q = heap.NewLPHeap(s.maxPerInt)
		s.queues[signal.Interval] = q
	}

	// Add signal to the interval queue
	q.Add(New(signal))
}

// Peek returns the highest priority signal for the specified interval
// without removing it from the queue. Returns nil if no signals exist
// for the interval. This method is thread-safe.
func (s *SignalService) Peek(interval string) *models.Signal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if q, ok := s.queues[interval]; ok {
		if item := q.Peek(); item != nil {
			return Extract(item.(*Signal))

		}
	}
	return nil
}

// Pop removes and returns the highest priority signal for the specified
// interval. Returns nil if no signals exist for the interval.
// This method is thread-safe.
func (s *SignalService) Pop(interval string) *models.Signal {
	s.mu.Lock()
	defer s.mu.Unlock()

	if q, ok := s.queues[interval]; ok {
		if item := q.Pop(); item != nil {
			return Extract(item.(*Signal))
		}
	}
	return nil
}

// Intervals returns a slice of all intervals that currently have signals.
// This is useful for downstream services to know which timeframes have
// available signals for processing. This method is thread-safe.
func (s *SignalService) Intervals() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, 0, len(s.queues))
	for interval, queue := range s.queues {
		if !queue.IsEmpty() {
			out = append(out, interval)
		}
	}
	return out
}

// Stats returns statistics about the signal service including the number
// of signals per interval. This is useful for monitoring and debugging.
func (s *SignalService) Stats(intervals ...string) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for _, interval := range intervals {
		queue, ok := s.queues[interval]
		if !ok {
			continue
		}
		stats[interval] = queue.Size()
	}
	return stats
}

// GetAvailable returns the number of available signals for an interval
func (s *SignalService) GetAvailable(intervals ...string) map[string][]*models.Signal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string][]*models.Signal)
	for _, interval := range intervals {
		queue, ok := s.queues[interval]
		if !ok {
			continue
		}
		out[interval] = make([]*models.Signal, 0, queue.Size())
		for _, item := range queue.Items() {
			out[interval] = append(out[interval], Extract(item.(*Signal)))
		}
	}
	return out

}
