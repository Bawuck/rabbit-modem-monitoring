package pages

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/connection"
	"example.com/4g-monitor/internal/model"
)

type Overview struct {
	Decorations       widget.Decorations
	list              widget.List
	settings          Settings
	configure         widget.Clickable
	initialized       bool
	settingsRequested bool
}

// RequestSettings is called only by this dashboard's window event loop.
func (p *Overview) RequestSettings() { p.settingsRequested = true }

// Layout keeps the custom title bar visible above both dashboard and settings.
func (p *Overview) Layout(gtx layout.Context, t *material.Theme, s model.Snapshot, controller *connection.Controller) layout.Dimensions {
	// Clickable.Layout consumes pending clicks, so handle navigation before
	// rendering the header button (not later in layoutContent).
	view := controller.Snapshot()
	if p.settingsRequested && !view.Busy {
		p.settingsRequested = false
		if !p.settings.open {
			p.settings.Open(view)
		}
	}
	if !p.initialized && !view.Busy {
		p.initialized = true
		if !view.Ready {
			p.settings.Open(view)
		}
	}
	for p.configure.Clicked(gtx) {
		if !view.Busy && !p.settings.open {
			p.settings.Open(view)
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 20, Right: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return p.Decorations.LayoutMove(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 8, Bottom: 8, Right: 20}.Layout(gtx, components.Label(t, 18, "Rabbit Modem Monitoring", components.Foreground, true))
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if controller.Snapshot().Busy || p.settings.open {
							gtx = gtx.Disabled()
						}
						return material.Button(t, &p.configure, "Pengaturan").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: 12}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						// Gio supplies vector controls and native action hit regions.
						gtx.Constraints.Min.X = gtx.Dp(140)
						gtx.Constraints.Max.X = gtx.Dp(140)
						style := material.Decorations(t, &p.Decorations,
							system.ActionMinimize|system.ActionMaximize|system.ActionClose, "")
						style.Background, style.Foreground = components.Background, components.Foreground
						return style.Layout(gtx)
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return p.layoutContent(gtx, t, s, controller) }),
	)
}

func NewOverview() *Overview {
	return &Overview{
		list: widget.List{List: layout.List{Axis: layout.Vertical}},
	}
}

func (p *Overview) layoutContent(gtx layout.Context, t *material.Theme, s model.Snapshot, controller *connection.Controller) layout.Dimensions {
	view := controller.Snapshot()
	if p.settings.open {
		dimensions := p.settings.Layout(gtx, t, controller, view)
		if p.settings.open {
			return dimensions
		}
	}
	host := s.Host
	if host == "" {
		host = "Koneksi belum diatur"
	}
	return layout.UniformInset(20).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		rows := []layout.Widget{
			func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return components.StatusBadge(gtx, t, s) }),
					layout.Rigid(layout.Spacer{Width: 10}.Layout),
					layout.Flexed(1, components.Label(t, 11, components.StateMessage(s), components.Muted, false)),
				)
			},
			func(gtx layout.Context) layout.Dimensions {
				return components.Row(gtx, 12,
					func(gtx layout.Context) layout.Dimensions {
						return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.SignalStrength(gtx, t, s, false) })
					},
					func(gtx layout.Context) layout.Dimensions {
						return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.NetworkInfo(gtx, t, s.Reading, false) })
					},
				)
			},
			func(gtx layout.Context) layout.Dimensions { return p.metrics(gtx, t, s) },
			func(gtx layout.Context) layout.Dimensions {
				return components.Card(gtx, func(gtx layout.Context) layout.Dimensions {
					return components.Column(gtx, 8,
						components.Label(t, 11, "CURRENT MODEM TRAFFIC · NOT A SPEED TEST", components.Muted, true),
						func(gtx layout.Context) layout.Dimensions {
							return components.ConnectionStats(gtx, t, s.Reading, false)
						},
					)
				})
			},
			func(gtx layout.Context) layout.Dimensions { return components.ModemDetails(gtx, t, s) },
			func(gtx layout.Context) layout.Dimensions { return p.charts(gtx, t, s) },
			components.Label(t, 11, components.UpdateText(s, gtx.Now), components.Muted, false),
			components.Label(t, 11, host+" · RSRQ/SINR units unverified: raw values shown.", components.Muted, false),
		}
		return material.List(t, &p.list).Layout(gtx, len(rows), func(gtx layout.Context, index int) layout.Dimensions {
			return (layout.Inset{Bottom: 12, Right: 8}).Layout(gtx, rows[index])
		})
	})
}

func (p *Overview) metrics(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	names := []string{"RSRP", "RSRQ", "SINR", "RSSI"}
	units := []string{"dBm", "raw", "raw", "dBm"}
	values := []model.Value[float64]{s.Reading.RSRP, s.Reading.RSRQ, s.Reading.SINR, s.Reading.RSSI}
	var columns []layout.Widget
	for i := range names {
		columns = append(columns, func(gtx layout.Context) layout.Dimensions {
			return components.Card(gtx, func(gtx layout.Context) layout.Dimensions {
				return components.SignalMetric(gtx, t, names[i], units[i], values[i], false)
			})
		})
	}
	return components.Row(gtx, 10, columns...)
}

func (p *Overview) charts(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	var charts []layout.Widget
	for _, metric := range []components.ChartMetric{components.ChartRSRP, components.ChartRSRQ, components.ChartSINR} {
		charts = append(charts, func(gtx layout.Context) layout.Dimensions {
			return components.SignalChart(gtx, t, s, metric)
		})
	}
	if gtx.Constraints.Max.X < gtx.Sp(unit.Sp(590)) {
		return components.Column(gtx, 12, charts...)
	}
	return components.Row(gtx, 10, charts...)
}
