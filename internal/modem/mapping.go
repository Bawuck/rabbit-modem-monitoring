package modem

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"example.com/4g-monitor/internal/model"
)

// scalar tolerates optional numbers encoded as strings or JSON numbers.
// Invalid optional values are unavailable, not zero and not a failed cycle.
type scalar struct {
	text     string
	valid    bool
	isString bool
}

func (s *scalar) UnmarshalJSON(data []byte) error {
	*s = scalar{}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		return nil
	}
	var text string
	if data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		var number json.Number
		if err := json.Unmarshal(data, &number); err != nil {
			return nil
		}
		text = number.String()
	}
	s.text, s.valid, s.isString = strings.TrimSpace(text), true, data[0] == '"'
	return nil
}

type monitorResponse struct {
	LoginInfo         scalar `json:"loginfo"`
	Operator          scalar `json:"network_provider"`
	Roaming           scalar `json:"simcard_roam"`
	StationCount      scalar `json:"sta_count"`
	MultiStationCount scalar `json:"m_sta_count"`
	TotalDownload     scalar `json:"realtime_rx_bytes"`
	TotalUpload       scalar `json:"realtime_tx_bytes"`
	ConnectionTime    scalar `json:"realtime_time"`
	Network           scalar `json:"network_type"`
	SubNetwork        scalar `json:"sub_network_type"`
	RSSI              scalar `json:"rssi"`
	RSRP              scalar `json:"lte_rsrp"`
	RSRQ              scalar `json:"nv_rsrq"`
	SINR              scalar `json:"nv_sinr"`
	Band              scalar `json:"lte_band"`
	PCI               scalar `json:"nv_pci"`
	PPP               scalar `json:"ppp_status"`
	Bars              scalar `json:"signalbar"`
	Download          scalar `json:"realtime_rx_thrpt"`
	Upload            scalar `json:"realtime_tx_thrpt"`
}

func counter(s scalar) model.Value[uint64] {
	if !s.valid || s.text == "" {
		return model.Value[uint64]{}
	}
	n, err := strconv.ParseUint(s.text, 10, 64)
	if err != nil {
		return model.Value[uint64]{}
	}
	return model.Some(n)
}

func optionalText(s scalar) model.Value[string] {
	if !s.valid || !s.isString || s.text == "" {
		return model.Value[string]{}
	}
	// Bound device-provided labels and remove control characters before display.
	text := []rune(strings.Map(func(r rune) rune {
		if r < 32 || (r >= 127 && r <= 159) {
			return -1
		}
		return r
	}, s.text))
	if len(text) > 80 {
		text = text[:80]
	}
	value := strings.TrimSpace(string(text))
	if value == "" {
		return model.Value[string]{}
	}
	return model.Some(value)
}

func number(s scalar) model.Value[float64] {
	if !s.valid || s.text == "" {
		return model.Value[float64]{}
	}
	n, err := strconv.ParseFloat(s.text, 64)
	if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return model.Value[float64]{}
	}
	return model.Some(n)
}

func integer(s scalar, upper int) model.Value[int] {
	n := number(s)
	if !n.Valid || n.Value < 0 || n.Value > float64(upper) || math.Trunc(n.Value) != n.Value {
		return model.Value[int]{}
	}
	return model.Some(int(n.Value))
}

func mbps(s scalar) model.Value[float64] {
	n := number(s)
	if !n.Valid || n.Value < 0 {
		return model.Value[float64]{}
	}
	// Divide first to avoid overflow for large finite raw values.
	return model.Some(n.Value / 1_000_000 * 8)
}

func normalized(value string) string {
	return strings.NewReplacer("-", "_", " ", "_").Replace(strings.ToUpper(strings.TrimSpace(value)))
}

func unavailableNetwork(value string) bool {
	switch normalized(value) {
	case "", "NO_SERVICE", "NO_SIGNAL", "LIMITED_SERVICE", "NONE":
		return true
	}
	return false
}

func networkName(network, subNetwork string) string {
	for _, value := range []string{subNetwork, network} {
		switch normalized(value) {
		case "LTE_A", "LTE+", "LTE_ADVANCED", "4G+":
			return "LTE-A" // Only explicit modem values may claim LTE-A.
		}
	}
	for _, value := range []string{network, subNetwork} {
		switch normalized(value) {
		case "LTE", "4G", "FDD", "TDD", "FDD_LTE", "TDD_LTE":
			return "LTE"
		}
	}
	return network
}

func mapResponse(response monitorResponse, at time.Time) model.Update {
	if !response.Network.valid || !response.Network.isString || !response.PPP.valid || !response.PPP.isString || response.PPP.text == "" {
		return model.Update{State: model.APIError, Message: "Required network/PPP status is missing or invalid"}
	}
	ppp := strings.ToLower(response.PPP.text)
	switch ppp {
	case "ppp_connected", "ppp_disconnected", "ppp_connecting", "ppp_disconnecting":
	default:
		return model.Update{State: model.APIError, Message: "Unrecognized PPP status from modem"}
	}
	if unavailableNetwork(response.Network.text) || (response.SubNetwork.valid && response.SubNetwork.text != "" && unavailableNetwork(response.SubNetwork.text)) {
		return model.Update{State: model.NoSignal, Message: "No service or limited service · measurements unavailable"}
	}
	if ppp != "ppp_connected" {
		return model.Update{State: model.Disconnected, Message: "Modem reachable · mobile data is not connected"}
	}
	network := networkName(response.Network.text, response.SubNetwork.text)
	reading := model.Reading{
		Operator: optionalText(response.Operator), Roaming: optionalText(response.Roaming),
		StationCount: counter(response.StationCount), MultiStationCount: counter(response.MultiStationCount),
		TotalDownload: counter(response.TotalDownload), TotalUpload: counter(response.TotalUpload),
		ConnectionTime: counter(response.ConnectionTime),
		Network:        model.Some(network), SignalBars: integer(response.Bars, 5),
		RSSI: number(response.RSSI), Download: mbps(response.Download), Upload: mbps(response.Upload),
	}
	if network == "LTE" || network == "LTE-A" {
		reading.RSRP, reading.RSRQ, reading.SINR = number(response.RSRP), number(response.RSRQ), number(response.SINR)
		reading.PCI = integer(response.PCI, 503)
		band := response.Band
		band.text = strings.TrimPrefix(strings.ToUpper(band.text), "B")
		if value := integer(band, 256); value.Valid && value.Value > 0 {
			reading.Band = model.Some("B" + strconv.Itoa(value.Value))
		}
	}
	return model.Update{State: model.Online, Reading: reading, ReceivedAt: at,
		Message: "Live modem traffic · polling every 2 seconds"}
}
