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
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return Badge(gtx, t, "Live", Accent) }),
	)
}

func stateColor(state model.ConnectionState) color.NRGBA {
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

func SignalStrength(gtx layout.Context, t *material.Theme, s model.Snapshot, compact bool) layout.Dimensions {
	col := Accent
	value, detail := "—", "Modem bars"
	if s.Reading.SignalBars.Valid {
		value = Integer(s.Reading.SignalBars) + "/5"
	}
	if s.State == model.Loading {
		detail = "Loading…"
	}
	if s.Stale {
		col = Amber
		detail += " · stale"
	}
	size := unit.Sp(38)
	if compact {
		size = 32
	}
	return Column(gtx, 4,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(Label(t, size, value, col, true)),
				layout.Rigid(layout.Spacer{Width: 10}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return Column(gtx, 2, Label(t, 11, "SIGNAL STRENGTH", Muted, false), Label(t, 12, detail, col, true))
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(4)))
			gap := gtx.Dp(4)
			width := max(0, size.X-4*gap)
			for i := 0; i < 5; i++ {
				fill := Border
				if s.Reading.SignalBars.Valid && i < s.Reading.SignalBars.Value {
					fill = col
				}
				left, right := width*i/5+gap*i, width*(i+1)/5+gap*i
				paint.FillShape(gtx.Ops, fill, clip.UniformRRect(image.Rect(left, 0, right, size.Y), gtx.Dp(2)).Op(gtx.Ops))
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
		Label(t, size, Number(value, -1), Foreground, true),
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
	if compact {
		size = 16
	}
	stat := func(title, value, unitName string) layout.Widget {
		return func(gtx layout.Context) layout.Dimensions {
			return Column(gtx, 2, Label(t, 10, title, Muted, false), Label(t, size, value, Foreground, true), Label(t, 10, unitName, Muted, false))
		}
	}
	return Row(gtx, 8,
		stat("DOWNLOAD", Number(r.Download, 3), "Mbps"),
		stat("UPLOAD", Number(r.Upload, 3), "Mbps"),
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
	if s.Message != "" {
		return s.Message
	}
	switch s.State {
	case model.Loading:
		return "Connecting to modem…"
	case model.NoSignal:
		return "No signal · measurements unavailable"
	case model.APIError:
		return "Modem API unavailable · retrying automatically"
	case model.Disconnected:
		return "Modem disconnected · retrying automatically"
	default:
		return "Live modem traffic · polling every 2 seconds"
	}
}
