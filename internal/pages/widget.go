// Package pages composes components into the widget and Overview dashboard.
package pages

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/model"
)

type Widget struct {
	Open  func()
	click widget.Clickable
}

func (p *Widget) Layout(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	for p.click.Clicked(gtx) {
		if p.Open != nil {
			p.Open()
		}
	}
	return material.Clickable(gtx, &p.click, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(12).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return components.Column(gtx, 6,
						func(gtx layout.Context) layout.Dimensions { return components.Header(gtx, t, "4G Monitor") },
						func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, components.Label(t, 12, components.Text(s.Reading.Network), components.Accent, true)),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions { return components.StatusBadge(gtx, t, s) }),
							)
						},
						func(gtx layout.Context) layout.Dimensions {
							return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.SignalScore(gtx, t, s, true) })
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
												return components.SignalMetric(gtx, t, "RSRQ", "dB", s.Reading.RSRQ, true)
											},
											func(gtx layout.Context) layout.Dimensions {
												return components.SignalMetric(gtx, t, "SINR", "dB", s.Reading.SINR, true)
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
						components.Label(t, 10, components.UpdateText(s, gtx.Now), components.Muted, false),
					)
				}),
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(components.Label(t, 11, "Open Overview →", components.Accent, true)),
			)
		})
	})
}
