package components

import (
	"fmt"
	"strconv"

	"gioui.org/layout"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/model"
)

func counterText(value model.Value[uint64]) string {
	if !value.Valid {
		return "—"
	}
	return strconv.FormatUint(value.Value, 10)
}

func trafficTotal(value model.Value[uint64]) string {
	if !value.Valid {
		return "—"
	}
	return fmt.Sprintf("%.2f GB (%s bytes)", float64(value.Value)/1_000_000_000, counterText(value))
}

// ModemDetails uses only snapshot values, so cached counters do not advance
// while stale. Device counters are not summed without verified semantics.
func ModemDetails(gtx layout.Context, t *material.Theme, s model.Snapshot) layout.Dimensions {
	r := s.Reading
	heading := "MODEM DETAILS"
	if s.Stale {
		heading += " · stale"
	}
	rows := []layout.Widget{Label(t, 11, heading, Muted, true)}
	for _, field := range []struct{ name, value string }{
		{"Operator", Text(r.Operator)},
		{"Roaming", Text(r.Roaming)},
		{"Device counter · sta_count", counterText(r.StationCount)},
		{"Device counter · m_sta_count", counterText(r.MultiStationCount)},
		{"Total download · modem counter", trafficTotal(r.TotalDownload)},
		{"Total upload · modem counter", trafficTotal(r.TotalUpload)},
		{"Connection time · realtime_time (raw)", counterText(r.ConnectionTime)},
	} {
		rows = append(rows, Label(t, 12, field.name+": "+field.value, Foreground, false))
	}
	rows = append(rows, Label(t, 11,
		"Device counter grouping and time units are unverified. Traffic reset period is modem-defined; totals are not remaining quota. GB = 1,000,000,000 bytes.", Muted, false))
	return Card(gtx, func(gtx layout.Context) layout.Dimensions {
		return Column(gtx, 8, rows...)
	})
}
