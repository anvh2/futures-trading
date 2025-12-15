package heap

import (
	"container/heap"
)

type Item interface {
	ID() string
	Score() float64
}

// Node represents an item in the item heap with its position index
type Node struct {
	item  Item
	index int
}

// Heap implements heap.Interface for item objects.
// This is a max heap where items with higher scores have higher priority.
type Heap []*Node

// Len returns the number of items in the heap
func (h Heap) Len() int {
	return len(h)
}

// Less compares two items for heap ordering.
// Returns true if item i has higher priority than item j.
// This creates a max heap where higher scores have higher priority.
func (h Heap) Less(i, j int) bool {
	return h[i].item.Score() > h[j].item.Score()
}

// Swap exchanges two items in the heap and updates their indices
func (h Heap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

// Push adds an item to the heap (used by heap.Push)
func (h *Heap) Push(x any) {
	item := x.(*Node)
	item.index = len(*h)
	*h = append(*h, item)
}

// Pop removes and returns the last item from the heap (used by heap.Pop)
func (h *Heap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // mark as removed
	*h = old[:n-1]
	return item
}

// LPHeap manages items using a priority heap where higher scoring items have higher priority.
// It maintains a fixed maximum size and automatically evicts lower priority items.
type LPHeap struct {
	maxSize int              // maximum number of items to keep
	items   map[string]*Node // id -> item mapping for O(1) lookups
	heap    Heap             // priority heap of items
}

// NewLPHeap creates a new heap with the specified maximum size.
// When the queue reaches maxSize, adding new items will evict the lowest priority item.
func NewLPHeap(maxSize int) *LPHeap {
	h := make(Heap, 0)
	heap.Init(&h)

	return &LPHeap{
		maxSize: maxSize,
		items:   make(map[string]*Node),
		heap:    h,
	}
}

// Add inserts or updates a item in the heap.
// If a item for the same id already exists, it updates the existing item.
// If the heap is at capacity, it removes the lowest priority item.
// items with invalid scores (≤ 0) are ignored.
func (q *LPHeap) Add(item Item) {
	// Ignore invalid scores
	if item.Score() <= 0 {
		return
	}

	// Check if item for this id already exists
	if existing, ok := q.items[item.ID()]; ok {
		// Update existing item and fix heap order
		existing.item = item
		heap.Fix(&q.heap, existing.index)
		return
	}

	// If at capacity, remove the lowest priority item
	if q.heap.Len() >= q.maxSize {
		removed := heap.Pop(&q.heap).(*Node)
		delete(q.items, removed.item.ID())
	}

	// Add new node
	node := &Node{item: item}
	heap.Push(&q.heap, node)
	q.items[item.ID()] = node
}

// Peek returns the highest priority item without removing it.
// Returns nil if the heap is empty.
func (q *LPHeap) Peek() Item {
	if q.heap.Len() == 0 {
		return nil
	}
	return q.heap[0].item
}

// Pop removes and returns the highest priority item.
// Returns nil if the heap is empty.
func (q *LPHeap) Pop() Item {
	if q.heap.Len() == 0 {
		return nil
	}

	item := heap.Pop(&q.heap).(*Node)
	delete(q.items, item.item.ID())
	return item.item
}

// Size returns the current number of items in the heap
func (q *LPHeap) Size() int {
	return q.heap.Len()
}

// Items returns all items in the heap
func (q *LPHeap) Items() []Item {
	out := make([]Item, 0, q.heap.Len())
	for _, item := range q.heap {
		out = append(out, item.item)
	}
	return out
}

// IsEmpty returns true if the heap contains no items
func (q *LPHeap) IsEmpty() bool {
	return q.heap.Len() == 0
}
