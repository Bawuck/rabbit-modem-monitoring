// Package connection serializes configuration persistence and worker replacement.
package connection

import (
	"context"
	"sync"

	"example.com/4g-monitor/internal/config"
	"example.com/4g-monitor/internal/model"
	"example.com/4g-monitor/internal/modem"
	"example.com/4g-monitor/internal/monitor"
	"example.com/4g-monitor/internal/startup"
)

// View is private application configuration, never a telemetry snapshot or log.
type View struct {
	StartupEnabled bool
	StartupBusy    bool
	StartupError   string
	Config         config.ConnectionConfig
	Ready          bool
	Busy           bool
	Error          string
}

type Controller struct {
	mu              sync.Mutex
	view            View
	requests        chan config.ConnectionConfig
	changed         chan struct{}
	startupRequests chan bool
}

func New() *Controller {
	return &Controller{view: View{Busy: true}, requests: make(chan config.ConnectionConfig, 1), changed: make(chan struct{}, 1), startupRequests: make(chan bool, 1)}
}

// SetStartup applies independently of the connection form and never restarts polling.
func (c *Controller) SetStartup(enabled bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.view.Busy || c.view.StartupBusy {
		return false
	}
	c.view.StartupBusy, c.view.StartupError = true, ""
	c.startupRequests <- enabled
	c.notify()
	return true
}

func (c *Controller) Snapshot() View           { c.mu.Lock(); defer c.mu.Unlock(); return c.view }
func (c *Controller) Changed() <-chan struct{} { return c.changed }
func (c *Controller) notify() {
	select {
	case c.changed <- struct{}{}:
	default:
	}
}

// Submit does no I/O and returns immediately to the window event loop.
func (c *Controller) Submit(value config.ConnectionConfig) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.view.Busy {
		return false
	}
	c.view.Busy, c.view.Error = true, ""
	c.requests <- value
	c.notify()
	return true
}

func (c *Controller) Run(ctx context.Context, store *monitor.Store) {
	var stop context.CancelFunc
	var done chan struct{}
	join := func() {
		if stop != nil {
			stop()
			<-done
			stop = nil
		}
	}
	defer join()
	start := func(value config.ConnectionConfig) {
		workerCtx, cancel := context.WithCancel(ctx)
		stop = cancel
		done = make(chan struct{})
		workerDone := done
		go func() { defer close(workerDone); modem.NewClient(value).Run(workerCtx, store.Apply) }()
	}
	value, err := config.Load()
	startupEnabled, startupErr := startup.Status()
	c.mu.Lock()
	c.view.StartupEnabled = startupEnabled
	if startupErr != nil {
		c.view.StartupError = startupErr.Error()
	}
	c.view.Busy = false
	if err != nil {
		c.view.Error = err.Error()
	}
	if err == nil && value.BaseURL != "" {
		c.view.Config, c.view.Ready = value, true
	}
	ready := c.view.Ready
	c.mu.Unlock()
	if ready {
		store.Reset(model.Loading, "Menghubungkan ke modem…", value.BaseURL)
		start(value)
	} else {
		store.Reset(model.Unconfigured, "Koneksi belum diatur", "")
	}
	c.notify()
	for {
		select {
		case <-ctx.Done():
			return
		case enabled := <-c.startupRequests:
			if ctx.Err() != nil {
				return
			}
			err := startup.Set(enabled)
			c.mu.Lock()
			c.view.StartupBusy = false
			if err != nil {
				c.view.StartupError = err.Error()
			} else {
				c.view.StartupEnabled, c.view.StartupError = enabled, ""
			}
			c.mu.Unlock()
			c.notify()
		case candidate := <-c.requests:
			if ctx.Err() != nil {
				return
			}
			validated, err := config.Validate(candidate)
			if err == nil {
				err = config.Save(validated)
			}
			if err == nil {
				join()
				if ctx.Err() != nil {
					return
				}
				store.Reset(model.Loading, "Menghubungkan ke modem…", validated.BaseURL)
				start(validated)
			}
			c.mu.Lock()
			c.view.Busy = false
			if err != nil {
				c.view.Error = err.Error()
			} else {
				c.view.Config, c.view.Ready, c.view.Error = validated, true, ""
			}
			c.mu.Unlock()
			c.notify()
		}
	}
}
