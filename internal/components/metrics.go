package components

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/model"
)

func Header(gtx layout.Context, t *material.Theme, title string) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, Label(t, 18, title, Foreground, true)),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return Badge(gtx, t, "Demo", Accent) }),
	)
}

func stateColor(state model.Scenario) color.NRGBA {
	switch state {
	case model.Online:
		return Green
	case model.APIError, model.NoSignal:
		return Red
	case model.Disconnected:
		return Amber
	default:
		return Blue
	}
}

func StatusBadge(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	value := string(s.State)
	if s.Stale {
		value += " · stale"
	}
	return Badge(gtx, t, value, stateColor(s.State))
}

func Number(v model.Value[float64], decimals int) string {
	if !v.Valid {
		return "—"
	}
	return strconv.FormatFloat(v.Value, 'f', decimals, 64)
}

func Integer(v model.Value[int]) string {
	if !v.Valid {
		return "—"
	}
	return strconv.Itoa(v.Value)
}

func Text(v model.Value[string]) string {
	if !v.Valid {
		return "—"
	}
	return v.Value
}

func qualityColor(quality string) color.NRGBA {
	switch quality {
	case "EXCELLENT":
		return Green
	case "GOOD":
		return Accent
	case "FAIR":
		return Amber
	case "POOR":
		return Red
	default:
		return Muted
	}
}

func SignalScore(gtx layout.Context, t *material.Theme, s model.Snapshot, compact bool) layout.Dimensions {
	col := qualityColor(s.Reading.Quality)
	quality := s.Reading.Quality
	if quality == "" {
		quality = "—"
	}
	if s.State == model.Loading {
		quality = "Loading…"
	}
	if s.Stale {
		col = Amber
		quality += " · stale"
	}
	size := unit.Sp(38)
	if compact {
		size = 32
	}
	return Column(gtx, 4,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(Label(t, size, Integer(s.Reading.Score), col, true)),
				layout.Rigid(layout.Spacer{Width: 10}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return Column(gtx, 2, Label(t, 11, "SIGNAL SCORE / 100", Muted, false), Label(t, 12, quality, col, true))
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(4)))
			paint.FillShape(gtx.Ops, Border, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(2)).Op(gtx.Ops))
			if s.Reading.Score.Valid {
				fill := size
				fill.X = size.X * max(0, min(100, s.Reading.Score.Value)) / 100
				paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: fill}, gtx.Dp(2)).Op(gtx.Ops))
			}
			return layout.Dimensions{Size: size}
		},
	)
}

func SignalMetric(gtx layout.Context, t *material.Theme, name, unitName string, value model.Value[float64], compact bool) layout.Dimensions {
	size := unit.Sp(24)
	if compact {
		size = 20
	}
	return Column(gtx, 1,
		Label(t, 11, name, Muted, true),
		Label(t, size, Number(value, 0), Foreground, true),
		Label(t, 10, unitName, Muted, false),
	)
}

func NetworkInfo(gtx layout.Context, t *material.Theme, r model.Reading, compact bool) layout.Dimensions {
	if compact {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, Label(t, 12, "Band  "+Text(r.Band), Foreground, false)),
			layout.Rigid(Label(t, 12, "PCI  "+Integer(r.PCI), Foreground, false)),
		)
	}
	return Column(gtx, 6,
		Label(t, 11, "NETWORK", Muted, true),
		Label(t, 21, Text(r.Network), Foreground, true),
		Label(t, 12, "Band "+Text(r.Band)+"  ·  PCI "+Integer(r.PCI), Muted, false),
	)
}

func ConnectionStats(gtx layout.Context, t *material.Theme, r model.Reading, compact bool) layout.Dimensions {
	size := unit.Sp(21)
	unitSuffix := " Mbps"
	if compact {
		size = 16
		unitSuffix = " M"
	}
	stat := func(title, value string) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			return Column(gtx, 3, Label(t, 10, title, Muted, false), Label(t, size, value, Foreground, true))
		}
	}
	return Row(gtx, 8,
		stat("DOWNLOAD", Number(r.Download, 1)+unitSuffix),
		stat("UPLOAD", Number(r.Upload, 1)+unitSuffix),
		stat("PING", Number(r.Ping, 0)+" ms"),
	)
}

func UpdateText(s model.Snapshot, now time.Time) string {
	if s.UpdatedAt.IsZero() {
		return "Waiting for data"
	}
	age := max(0, int(now.Sub(s.UpdatedAt).Seconds()))
	value := fmt.Sprintf("Updated %s · %ds ago", s.UpdatedAt.Format("15:04:05"), age)
	if s.Stale {
		value += " · stale"
	}
	return value
}

func StateMessage(s model.Snapshot) string {
	switch s.State {
	case model.Loading:
		return "Waiting for demo measurements…"
	case model.NoSignal:
		return "No signal · measurements unavailable"
	case model.APIError:
		return "Simulated API error · no live request"
	case model.Disconnected:
		return "Disconnected · demo updates paused"
	default:
		return "Simulated data · refreshes every 2 seconds"
	}
}
