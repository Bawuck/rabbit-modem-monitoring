// Package pages composes components into the widget and Overview dashboard.
package pages

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/model"
)

type Widget struct {
	Decorations   widget.Decorations
	Open          func()
	OpenSettings  func()
	click         widget.Clickable
	settingsClick widget.Clickable
	overviewClick widget.Clickable
}

func (p *Widget) Layout(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	for p.settingsClick.Clicked(gtx) {
		if p.OpenSettings != nil {
			p.OpenSettings()
		}
	}
	for p.overviewClick.Clicked(gtx) {
		if p.Open != nil {
			p.Open()
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			style := material.Decorations(t, &p.Decorations,
				system.ActionMinimize|system.ActionMaximize|system.ActionClose, "Rabbit Modem Monitoring")
			style.Background, style.Foreground = components.Background, components.Foreground
			style.Title.Color = components.Foreground
			style.Title.TextSize = 12
			return style.Layout(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return p.layoutContent(gtx, t, s) }),
		// Reserve timestamp space before allocating the remaining height to cards.
		// It must not compete with the footer inside a constrained body column.
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if s.State == model.Unconfigured {
				return layout.Dimensions{}
			}
			return layout.Inset{Left: 12, Right: 12, Top: 2, Bottom: 4}.Layout(gtx,
				components.Label(t, 10, components.UpdateText(s, gtx.Now), components.Muted, false))
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 12, Right: 12, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &p.overviewClick, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4}.Layout(gtx, components.Label(t, 11, "Open Overview →", components.Accent, true))
						})
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &p.settingsClick, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{Top: 4, Bottom: 4, Left: 8}.Layout(gtx, components.Label(t, 11, "Pengaturan", components.Accent, true))
						})
					}),
				)
			})
		}),
	)
}

// Only the body opens Overview; the title bar owns drag and close gestures.
func (p *Widget) layoutContent(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	for p.click.Clicked(gtx) {
		if p.Open != nil {
			p.Open()
		}
	}
	return material.Clickable(gtx, &p.click, func(gtx layout.Context) layout.Dimensions {
		if s.State == model.Unconfigured {
			return layout.UniformInset(12).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return components.Column(gtx, 12,
					components.Label(t, 16, "Koneksi belum diatur", components.Amber, true),
					components.Label(t, 12, "Klik untuk mengatur host dan password modem.", components.Muted, false),
				)
			})
		}
		return (layout.Inset{Left: 12, Right: 12, Top: 4, Bottom: 4}).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return components.Column(gtx, 4,
						func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, components.Label(t, 12, components.Text(s.Reading.Network), components.Accent, true)),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions { return components.StatusBadge(gtx, t, s) }),
							)
						},
						func(gtx layout.Context) layout.Dimensions {
							return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.SignalStrength(gtx, t, s, true) })
						},
						func(gtx layout.Context) layout.Dimensions {
							return components.Card(gtx, func(gtx layout.Context) layout.Dimensions {
								return components.Column(gtx, 8,
									func(gtx layout.Context) layout.Dimensions {
										return components.Row(gtx, 8,
											func(gtx layout.Context) layout.Dimensions {
												return components.SignalMetric(gtx, t, "RSRP", "dBm", s.Reading.RSRP, true)
											},
											func(gtx layout.Context) layout.Dimensions {
												return components.SignalMetric(gtx, t, "RSSI", "dBm", s.Reading.RSSI, true)
											},
											func(gtx layout.Context) layout.Dimensions {
												return components.SignalMetric(gtx, t, "RSRQ", "raw", s.Reading.RSRQ, true)
											},
											func(gtx layout.Context) layout.Dimensions {
												return components.SignalMetric(gtx, t, "SINR", "raw", s.Reading.SINR, true)
											},
										)
									},
									func(gtx layout.Context) layout.Dimensions { return components.NetworkInfo(gtx, t, s.Reading, true) },
								)
							})
						},
						func(gtx layout.Context) layout.Dimensions {
							return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.ConnectionStats(gtx, t, s.Reading, true) })
						},
					)
				}),
				layout.Flexed(1, layout.Spacer{}.Layout),
			)
		})
	})
}
