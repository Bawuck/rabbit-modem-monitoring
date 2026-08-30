package mock

import "example.com/4g-monitor/internal/model"

// Scores and quality labels are hand-authored demo fixtures, not calculations.
var fixtures = [...]model.Reading{
	reading("LTE", 82, "GOOD", -84, -9, 21, -57, "B3", 153, 12.4, 3.1, 28),
	reading("LTE-A", 91, "EXCELLENT", -77, -7, 27, -50, "B3 + B7", 153, 38.6, 8.2, 22),
	reading("LTE-A", 87, "GOOD", -81, -8, 24, -54, "B3 + B7", 153, 31.8, 6.7, 25),
	reading("LTE", 65, "FAIR", -98, -13, 11, -70, "B3", 153, 8.5, 2.2, 47),
	reading("LTE", 38, "POOR", -113, -18, 2, -85, "B8", 241, 1.8, 0.4, 96),
	reading("LTE", 74, "GOOD", -91, -11, 17, -63, "B3", 153, 10.6, 2.8, 34),
}

func reading(network string, score int, quality string, rsrp, rsrq, sinr, rssi float64,
	band string, pci int, down, up, ping float64) model.Reading {
	return model.Reading{
		Network: model.Some(network), Score: model.Some(score), Quality: quality,
		RSRP: model.Some(rsrp), RSRQ: model.Some(rsrq), SINR: model.Some(sinr), RSSI: model.Some(rssi),
		Band: model.Some(band), PCI: model.Some(pci), Download: model.Some(down),
		Upload: model.Some(up), Ping: model.Some(ping),
	}
}
