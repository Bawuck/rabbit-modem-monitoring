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
	"example.com/4g-monitor/internal/connection"
	"example.com/4g-monitor/internal/model"
	"example.com/4g-monitor/internal/monitor"
	"example.com/4g-monitor/internal/pages"
)

type windowHandle struct {
	window   *app.Window
	raise    chan struct{}
	settings chan struct{}
	close    chan struct{}
}

type closedWindow struct {
	handle *windowHandle
	err    error
}

func newHandle() *windowHandle {
	return &windowHandle{
		window:   new(app.Window),
		raise:    make(chan struct{}, 1),
		settings: make(chan struct{}, 1),
		close:    make(chan struct{}),
	}
}

// Run returns only after the windows, polling worker, and requests have stopped.
// main must call app.Main on the main goroutine.
func Run() error {
	store := monitor.NewStore()
	store.Reset(model.Unconfigured, "Koneksi belum diatur", "")
	controller := connection.New()
	ctx, cancel := context.WithCancel(context.Background())
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		controller.Run(ctx, store)
	}()
	defer func() {
		cancel()
		<-pollDone // No store/window lock is held while joining the worker.
	}()
	openDashboard := make(chan struct{}, 1)
	openSettings := make(chan struct{}, 1)
	closed := make(chan closedWindow, 2)
	widgetWindow := newHandle()
	var dashboard *windowHandle
	requestOpen := func() {
		select {
		case openDashboard <- struct{}{}:
		default: // Coalesce rapid clicks; there can only be one dashboard.
		}
	}
	requestSettings := func() {
		select {
		case openSettings <- struct{}{}:
		default:
		}
	}
	go widgetWindow.loop(true, store, controller, requestOpen, requestSettings, closed)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	closing := false
	var widgetErr error
	initialConfigHandled := false

	// Only this coordinator owns the window registry and closes command channels.
	for {
		select {
		case <-controller.Changed():
			if closing {
				continue
			}
			view := controller.Snapshot()
			if !initialConfigHandled && !view.Busy {
				initialConfigHandled = true
				if !view.Ready {
					requestOpen()
				}
			}
			widgetWindow.window.Invalidate()
			if dashboard != nil {
				dashboard.window.Invalidate()
			}
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
				go dashboard.loop(false, store, controller, nil, nil, closed)
			} else {
				select {
				case dashboard.raise <- struct{}{}:
				default:
				}
				dashboard.window.Invalidate()
			}
		case <-openSettings:
			if closing {
				continue
			}
			if dashboard == nil {
				dashboard = newHandle()
				go dashboard.loop(false, store, controller, nil, nil, closed)
			} else {
				select {
				case dashboard.raise <- struct{}{}:
				default:
				}
				dashboard.window.Invalidate()
			}
			select {
			case dashboard.settings <- struct{}{}:
			default:
			}
			dashboard.window.Invalidate()
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

func (h *windowHandle) loop(isWidget bool, store *monitor.Store, controller *connection.Controller, open func(), openSettings func(), closed chan<- closedWindow) {
	var err error
	defer func() { closed <- closedWindow{handle: h, err: err} }()
	if isWidget {
		h.window.Option(app.Decorated(false), app.Title("Rabbit Modem Monitoring · Live"), app.Size(unit.Dp(300), unit.Dp(380)),
			app.MinSize(unit.Dp(300), unit.Dp(380)), app.MaxSize(unit.Dp(300), unit.Dp(380)), app.TopMost(true))
	} else {
		h.window.Option(app.Decorated(false), app.Title("Rabbit Modem Monitoring · Overview · Live"), app.Size(unit.Dp(900), unit.Dp(650)),
			app.MinSize(unit.Dp(760), unit.Dp(520)))
	}
	theme := components.NewTheme()
	widgetPage := pages.Widget{Open: open, OpenSettings: openSettings}
	overview := pages.NewOverview()
	var ops op.Ops
	closing := false
	for {
		e := h.window.Event()
		if config, ok := e.(app.ConfigEvent); ok {
			if isWidget {
				widgetPage.Decorations.Maximized = config.Config.Mode == app.Maximized
			} else {
				overview.Decorations.Maximized = config.Config.Mode == app.Maximized
			}
		}
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
		if !isWidget && !closing {
			select {
			case <-h.settings:
				overview.RequestSettings()
			default:
			}
		}
		if frame, ok := e.(app.FrameEvent); ok {
			gtx := app.NewContext(&ops, frame)
			paint.Fill(gtx.Ops, components.Background)
			// Both pages use one detached, coherent snapshot per frame.
			snapshot := store.Snapshot()
			if isWidget {
				if actions := widgetPage.Decorations.Update(gtx); actions != 0 {
					h.window.Perform(actions)
				}
				widgetPage.Layout(gtx, theme, snapshot)
			} else {
				if actions := overview.Decorations.Update(gtx); actions != 0 {
					h.window.Perform(actions)
				}
				overview.Layout(gtx, theme, snapshot, controller)
			}
			frame.Frame(gtx.Ops)
		}
	}
}
