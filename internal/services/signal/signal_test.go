package signal

// import (
// 	"testing"
// 	"time"

// 	"github.com/anvh2/futures-trading/internal/libs/heap"
// 	"github.com/anvh2/futures-trading/internal/models"
// )

// func TestSignalService(t *testing.T) {
// 	// Create a new signal service
// 	svc := NewSignalService(10)

// 	// Create test oscillator data
// 	oscillator := &models.Oscillator{
// 		Symbol: "BTCUSDT",
// 		Stoch: map[string]*models.Stoch{
// 			"5m": {RSI: 75.5, K: 80.2, D: 78.1},
// 			"1h": {RSI: 68.3, K: 72.4, D: 70.8},
// 		},
// 	}

// 	// Test ingesting signals
// 	signal5m := NewRSISignal(oscillator, "5m")
// 	signal1h := NewRSISignal(oscillator, "1h")

// 	svc.Ingest(signal5m)
// 	svc.Ingest(signal1h)

// 	// Test that intervals are correctly tracked
// 	intervals := svc.Intervals()
// 	if len(intervals) != 2 {
// 		t.Errorf("Expected 2 intervals, got %d", len(intervals))
// 	}

// 	// Test peeking at signals
// 	best5m := svc.Peek("5m")
// 	if best5m == nil {
// 		t.Fatal("Expected signal for 5m interval")
// 	}
// 	if best5m.Symbol() != "BTCUSDT" {
// 		t.Errorf("Expected symbol BTCUSDT, got %s", best5m.Symbol())
// 	}
// 	if best5m.Score() != 75.5 {
// 		t.Errorf("Expected score 75.5, got %f", best5m.Score())
// 	}

// 	// Test that peek doesn't remove the signal
// 	stillThere := svc.Peek("5m")
// 	if stillThere == nil {
// 		t.Fatal("Signal should still be there after peek")
// 	}

// 	// Test popping signals
// 	popped := svc.Pop("5m")
// 	if popped == nil {
// 		t.Fatal("Expected to pop a signal")
// 	}
// 	if popped.Symbol() != "BTCUSDT" {
// 		t.Errorf("Expected popped symbol BTCUSDT, got %s", popped.Symbol())
// 	}

// 	// Test that signal is gone after pop
// 	empty := svc.Peek("5m")
// 	if empty != nil {
// 		t.Error("Signal should be gone after pop")
// 	}

// 	// Test stats
// 	stats := svc.Stats()
// 	if stats["1h"] != 1 {
// 		t.Errorf("Expected 1 signal in 1h interval, got %d", stats["1h"])
// 	}
// 	if stats["5m"] != 0 {
// 		t.Errorf("Expected 0 signals in 5m interval, got %d", stats["5m"])
// 	}
// }

// func TestLPHeap(t *testing.T) {
// 	queue := heap.NewLPHeap(3) // max 3 signals

// 	// Create test oscillators with different RSI scores
// 	osc1 := &models.Oscillator{Symbol: "BTC", Stoch: map[string]*models.Stoch{"5m": {RSI: 80}}}
// 	osc2 := &models.Oscillator{Symbol: "ETH", Stoch: map[string]*models.Stoch{"5m": {RSI: 70}}}
// 	osc3 := &models.Oscillator{Symbol: "ADA", Stoch: map[string]*models.Stoch{"5m": {RSI: 90}}}
// 	osc4 := &models.Oscillator{Symbol: "DOT", Stoch: map[string]*models.Stoch{"5m": {RSI: 60}}}

// 	// Add signals
// 	queue.Add(NewRSISignal(osc1, "5m")) // RSI 80
// 	queue.Add(NewRSISignal(osc2, "5m")) // RSI 70
// 	queue.Add(NewRSISignal(osc3, "5m")) // RSI 90

// 	// ADA should be at the top (highest RSI)
// 	top := queue.Peek().(Signal)
// 	if top.Symbol() != "ADA" {
// 		t.Errorf("Expected ADA at top, got %s", top.Symbol())
// 	}

// 	// Adding 4th signal should evict the lowest (ETH with RSI 70)
// 	queue.Add(NewRSISignal(osc4, "5m")) // RSI 60

// 	// Size should still be 3
// 	if queue.Size() != 3 {
// 		t.Errorf("Expected size 3, got %d", queue.Size())
// 	}

// 	// DOT (RSI 60) should NOT be in queue as it has lower RSI than ETH (70)
// 	// Actually, it should evict the lowest priority signal
// }

// func TestRSISignal(t *testing.T) {
// 	oscillator := &models.Oscillator{
// 		Symbol: "BTCUSDT",
// 		Stoch: map[string]*models.Stoch{
// 			"5m": {RSI: 75.5, K: 80.2, D: 78.1},
// 		},
// 	}

// 	signal := NewRSISignal(oscillator, "5m")

// 	// Test interface compliance
// 	if signal.Symbol() != "BTCUSDT" {
// 		t.Errorf("Expected symbol BTCUSDT, got %s", signal.Symbol())
// 	}

// 	if signal.Interval() != "5m" {
// 		t.Errorf("Expected interval 5m, got %s", signal.Interval())
// 	}

// 	if signal.Score() != 75.5 {
// 		t.Errorf("Expected score 75.5, got %f", signal.Score())
// 	}

// 	if signal.Payload() != oscillator {
// 		t.Error("Expected payload to be the original oscillator")
// 	}

// 	// Test timestamp is recent
// 	if time.Since(signal.Timestamp()) > time.Second {
// 		t.Error("Signal timestamp should be recent")
// 	}
// }
