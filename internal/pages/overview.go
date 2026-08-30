package pages

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/model"
)

type Overview struct {
	list      widget.List
	choices   []model.Scenario
	buttons   []widget.Clickable
	selectOne func(model.Scenario)
}

func NewOverview(selectOne func(model.Scenario)) *Overview {
	choices := model.Scenarios()
	return &Overview{
		list:    widget.List{List: layout.List{Axis: layout.Vertical}},
		choices: choices, buttons: make([]widget.Clickable, len(choices)), selectOne: selectOne,
	}
}

func (p *Overview) Update(gtx layout.Context) {
	for i := range p.buttons {
		for p.buttons[i].Clicked(gtx) {
			p.selectOne(p.choices[i])
		}
	}
}

func (p *Overview) Layout(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	return layout.UniformInset(20).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Dp(136)
				gtx.Constraints.Max.X = gtx.Dp(136)
				return p.sidebar(gtx, t)
			}),
			layout.Rigid(layout.Spacer{Width: 20}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				rows := []layout.Widget{
					func(gtx layout.Context) layout.Dimensions {
						return components.Column(gtx, 6,
							func(gtx layout.Context) layout.Dimensions { return components.Header(gtx, t, "Overview") },
							components.Label(t, 12, "Your connection, at a glance.", components.Muted, false),
						)
					},
					func(gtx layout.Context) layout.Dimensions { return p.scenarioPicker(gtx, t, s.State) },
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
								return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.SignalScore(gtx, t, s, false) })
							},
							func(gtx layout.Context) layout.Dimensions {
								return components.Card(gtx, func(gtx layout.Context) layout.Dimensions { return components.NetworkInfo(gtx, t, s.Reading, false) })
							},
						)
					},
					func(gtx layout.Context) layout.Dimensions { return p.metrics(gtx, t, s) },
					func(gtx layout.Context) layout.Dimensions {
						return components.Card(gtx, func(gtx layout.Context) layout.Dimensions {
							return components.ConnectionStats(gtx, t, s.Reading, false)
						})
					},
					func(gtx layout.Context) layout.Dimensions { return p.charts(gtx, t, s) },
					components.Label(t, 11, components.UpdateText(s, gtx.Now), components.Muted, false),
					components.Label(t, 11, "Demo only · no device connected · no speed test", components.Muted, false),
				}
				return material.List(t, &p.list).Layout(gtx, len(rows), func(gtx layout.Context, index int) layout.Dimensions {
					return (layout.Inset{Bottom: 12, Right: 8}).Layout(gtx, rows[index])
				})
			}),
		)
	})
}

func (p *Overview) sidebar(gtx layout.Context, t *material.Theme) layout.Dimensions {
	items := []layout.Widget{
		components.Label(t, 18, "4G Monitor", components.Foreground, true),
		components.Label(t, 10, "CONNECTION MONITOR", components.Muted, false),
		func(gtx layout.Context) layout.Dimensions {
			return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return components.Badge(gtx, t, "Demo", components.Accent)
			})
		},
		layout.Spacer{Height: 16}.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return components.Card(gtx, components.Label(t, 13, "Overview", components.Accent, true))
		},
	}
	for _, name := range []string{"Signal", "Cell", "History", "Settings"} {
		items = append(items, func(gtx layout.Context) layout.Dimensions {
			// No Clickable is registered: these are deliberately unavailable pages.
			return (layout.Inset{Left: 12, Top: 7, Bottom: 7}).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return components.Column(gtx, 3,
					components.Label(t, 13, name, components.Muted, false),
					components.Label(t, 10, "Coming soon", components.Muted, false),
				)
			})
		})
	}
	return components.Column(gtx, 10, items...)
}

func (p *Overview) scenarioPicker(gtx layout.Context, t *material.Theme, selected model.Scenario) layout.Dimensions {
	return components.Column(gtx, 7,
		components.Label(t, 10, "DEMO SCENARIO", components.Muted, true),
		func(gtx layout.Context) layout.Dimensions {
			// Use two rows at narrow widths or larger text scales.
			count := 5
			if gtx.Constraints.Max.X < gtx.Sp(unit.Sp(620)) {
				count = 3
			}
			var rows []layout.Widget
			for start := 0; start < len(p.choices); start += count {
				end := min(start+count, len(p.choices))
				rows = append(rows, func(gtx layout.Context) layout.Dimensions {
					var buttons []layout.Widget
					for i := start; i < end; i++ {
						buttons = append(buttons, func(gtx layout.Context) layout.Dimensions {
							b := material.Button(t, &p.buttons[i], string(p.choices[i]))
							b.TextSize = 11
							b.Inset = layout.Inset{Top: 8, Bottom: 8, Left: 5, Right: 5}
							b.CornerRadius = 7
							if selected != p.choices[i] {
								b.Background, b.Color = components.Surface, components.Foreground
							}
							return b.Layout(gtx)
						})
					}
					return components.Row(gtx, 6, buttons...)
				})
			}
			return components.Column(gtx, 6, rows...)
		},
	)
}

func (p *Overview) metrics(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	names := []string{"RSRP", "RSRQ", "SINR", "RSSI"}
	units := []string{"dBm", "dB", "dB", "dBm"}
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
