package components

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/model"
)

type ChartMetric int

const (
	ChartRSRP ChartMetric = iota
	ChartRSRQ
	ChartSINR
)

func (m ChartMetric) value(s model.Sample) model.Value[float64] {
	switch m {
	case ChartRSRQ:
		return s.RSRQ
	case ChartSINR:
		return s.SINR
	default:
		return s.RSRP
	}
}

func SignalChart(gtx layout.Context, t *material.Theme, s model.Snapshot, metric ChartMetric) layout.Dimensions {
	title, unitName, col := "RSRP", "dBm", Accent
	if metric == ChartRSRQ {
		title, unitName, col = "RSRQ", "raw", Blue
	} else if metric == ChartSINR {
		title, unitName, col = "SINR", "raw", Green
	}
	if s.Stale {
		title += " · stale"
		col = Amber
	}
	low, high := math.Inf(1), math.Inf(-1)
	for _, sample := range s.History {
		v := metric.value(sample)
		if v.Valid && !math.IsNaN(v.Value) && !math.IsInf(v.Value, 0) {
			low = math.Min(low, v.Value)
			high = math.Max(high, v.Value)
		}
	}
	hasValues := !math.IsInf(low, 1)
	if hasValues {
		padding := math.Max(1, (high-low)*0.15)
		low = math.Floor(low - padding)
		high = math.Ceil(high + padding)
	}
	rangeText, timeText := "No samples · "+unitName, "Waiting for data"
	if hasValues {
		rangeText = fmt.Sprintf("%.0f … %.0f %s", low, high, unitName)
		timeText = s.History[0].At.Format("15:04:05") + " — " + s.History[len(s.History)-1].At.Format("15:04:05")
	}
	return Card(gtx, func(gtx layout.Context) layout.Dimensions {
		return Column(gtx, 7,
			Label(t, 12, title, col, true),
			Label(t, 10, rangeText, Muted, false),
			func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(96)))
				defer clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops).Pop()
				for i := 0; i < 3; i++ {
					y := (size.Y - 1) * i / 2
					paint.FillShape(gtx.Ops, Border, clip.Rect(image.Rect(0, y, size.X, y+1)).Op())
				}
				if hasValues && size.X > 0 && size.Y > 0 {
					drawSeries(gtx, size, s.History, metric, low, high, col)
				} else {
					gtx.Constraints = layout.Exact(size)
					layout.Center.Layout(gtx, Label(t, 20, "—", Muted, false))
				}
				return layout.Dimensions{Size: size}
			},
			Label(t, 10, timeText, Muted, false),
		)
	})
}

func drawSeries(gtx layout.Context, size image.Point, samples []model.Sample, metric ChartMetric, low, high float64, col color.NRGBA) {
	inset := float32(gtx.Dp(4))
	width, height := max(0, float32(size.X)-2*inset), max(0, float32(size.Y)-2*inset)
	start := samples[0].At
	span := samples[len(samples)-1].At.Sub(start).Seconds()
	var path clip.Path
	path.Begin(gtx.Ops)
	connected := false
	points := make([]f32.Point, 0, len(samples))
	for _, sample := range samples {
		v := metric.value(sample)
		if !v.Valid || math.IsNaN(v.Value) || math.IsInf(v.Value, 0) {
			connected = false // Missing measurements break the line, never plot zero.
			continue
		}
		x := inset + width/2
		if span > 0 {
			x = inset + width*float32(sample.At.Sub(start).Seconds()/span)
		}
		y := inset + height*(1-float32((v.Value-low)/(high-low)))
		point := f32.Pt(x, y)
		if connected {
			path.LineTo(point)
		} else {
			path.MoveTo(point)
		}
		connected = true
		points = append(points, point)
	}
	paint.FillShape(gtx.Ops, col, clip.Stroke{Path: path.End(), Width: float32(gtx.Dp(2))}.Op())
	// Draw sample markers so a single sample or an isolated valid value is visible.
	radius := max(2, gtx.Dp(2))
	for _, point := range points {
		x, y := int(math.Round(float64(point.X))), int(math.Round(float64(point.Y)))
		paint.FillShape(gtx.Ops, col, clip.Ellipse(image.Rect(x-radius, y-radius, x+radius, y+radius)).Op(gtx.Ops))
	}
}
