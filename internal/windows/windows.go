// Package windows owns the process lifecycle and the two independent Gio windows.
package windows

import (
	"context"
	"log"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/modem"
	"example.com/4g-monitor/internal/monitor"
	"example.com/4g-monitor/internal/pages"
)

type windowHandle struct {
	window *app.Window
	raise  chan struct{}
	close  chan struct{}
}

type closedWindow struct {
	handle *windowHandle
	err    error
}

func newHandle() *windowHandle {
	return &windowHandle{
		window: new(app.Window),
		raise:  make(chan struct{}, 1),
		close:  make(chan struct{}),
	}
}

// Run returns only after the windows, polling worker, and requests have stopped.
// main must call app.Main on the main goroutine.
func Run() error {
	store := monitor.NewStore()
	ctx, cancel := context.WithCancel(context.Background())
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		modem.NewClient().Run(ctx, store.Apply)
	}()
	defer func() {
		cancel()
		<-pollDone // No store/window lock is held while joining the worker.
	}()
	openDashboard := make(chan struct{}, 1)
	closed := make(chan closedWindow, 2)
	widgetWindow := newHandle()
	var dashboard *windowHandle
	requestOpen := func() {
		select {
		case openDashboard <- struct{}{}:
		default: // Coalesce rapid clicks; there can only be one dashboard.
		}
	}
	go widgetWindow.loop(true, store, requestOpen, closed)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	closing := false
	var widgetErr error

	// Only this coordinator owns the window registry and closes command channels.
	for {
		select {
		case <-ticker.C:
			if closing {
				continue
			}
			// Repaint data age while requests are in flight or data is stale.
			// This ticker never writes a measurement or starts a request.
			widgetWindow.window.Invalidate()
			if dashboard != nil {
				dashboard.window.Invalidate()
			}
		case <-store.Changed():
			if !closing {
				widgetWindow.window.Invalidate()
				if dashboard != nil {
					dashboard.window.Invalidate()
				}
			}
		case <-openDashboard:
			if closing {
				continue
			}
			if dashboard == nil {
				dashboard = newHandle()
				go dashboard.loop(false, store, nil, closed)
			} else {
				select {
				case dashboard.raise <- struct{}{}:
				default:
				}
				dashboard.window.Invalidate()
			}
		case result := <-closed:
			if result.handle == widgetWindow {
				closing = true
				widgetErr = result.err
				cancel()
				ticker.Stop()
				if dashboard != nil {
					close(dashboard.close)
					dashboard.window.Invalidate()
				}
			} else if result.handle == dashboard {
				dashboard = nil
				if result.err != nil {
					log.Printf("dashboard: %v", result.err)
				}
			}
			if closing && dashboard == nil {
				return widgetErr
			}
		}
	}
}

func (h *windowHandle) loop(isWidget bool, store *monitor.Store, open func(), closed chan<- closedWindow) {
	var err error
	defer func() { closed <- closedWindow{handle: h, err: err} }()
	if isWidget {
		h.window.Option(app.Title("4G Monitor · Live"), app.Size(unit.Dp(300), unit.Dp(380)),
			app.MinSize(unit.Dp(300), unit.Dp(380)), app.MaxSize(unit.Dp(300), unit.Dp(380)), app.TopMost(true))
	} else {
		h.window.Option(app.Title("4G Monitor · Overview · Live"), app.Size(unit.Dp(900), unit.Dp(650)),
			app.MinSize(unit.Dp(760), unit.Dp(520)))
	}
	theme := components.NewTheme()
	widgetPage := pages.Widget{Open: open}
	overview := pages.NewOverview()
	var ops op.Ops
	closing := false
	for {
		e := h.window.Event()
		if destroyed, ok := e.(app.DestroyEvent); ok {
			err = destroyed.Err
			return
		}
		if !closing {
			select {
			case <-h.close:
				closing = true
				h.window.Perform(system.ActionClose)
			default:
			}
		}
		select {
		case <-h.raise:
			if !closing {
				h.window.Perform(system.ActionRaise)
			}
		default:
		}
		if frame, ok := e.(app.FrameEvent); ok {
			gtx := app.NewContext(&ops, frame)
			paint.Fill(gtx.Ops, components.Background)
			// Both pages use one detached, coherent snapshot per frame.
			snapshot := store.Snapshot()
			if isWidget {
				widgetPage.Layout(gtx, theme, snapshot)
			} else {
				overview.Layout(gtx, theme, snapshot)
			}
			frame.Frame(gtx.Ops)
		}
	}
}
