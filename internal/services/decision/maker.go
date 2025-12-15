package decision

import (
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/services/settings"
	"github.com/anvh2/futures-trading/internal/services/signal"
	"github.com/anvh2/futures-trading/internal/services/state"
)

// Maker interface defines the decision-making capabilities
type IMaker interface {
	Start() error // Start queue processing
	Stop()        // Stop processing
}

// Maker implementation
type Maker struct {
	logger      *logger.Logger
	state       *state.StateManager
	queue       *queue.Queue
	signal      signal.Service
	settings    *settings.Settings
	quitChannel chan struct{}
}

// NewMakNewer creates a new decision maker
func New(
	logger *logger.Logger,
	state *state.StateManager,
	queue *queue.Queue,
	signal signal.Service,
	settings *settings.Settings,
) IMaker {
	return &Maker{
		logger:      logger,
		state:       state,
		queue:       queue,
		signal:      signal,
		settings:    settings,
		quitChannel: make(chan struct{}),
	}
}

// Start begins the queue processing goroutine
func (de *Maker) Start() error {
	if err := de.HandleSignals(); err != nil {
		return err
	}

	return nil
}

// Stop stops the queue processing
func (de *Maker) Stop() {
	close(de.quitChannel)
}
