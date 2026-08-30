// Package modem reads the two fixed, local modem endpoints. It never writes
// device settings, logs response bodies, or persists device identifiers.
package modem

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"example.com/4g-monitor/internal/model"
)

// Keep the complete user-supplied URLs, including all query parameters.
const SignalURL = "http://192.168.100.1/goform/goform_get_cmd_process?cmd=network_type%2Csub_network_type%2Crssi%2Clte_rscp%2Clte_rsrp%2Cnv_rsrq%2Cnv_sinr%2Clte_band%2Ccell_id%2Cziccid%2Cnv_pci&multi_data=1"
const StatusURL = "http://192.168.100.1/goform/goform_get_cmd_process?multi_data=1&sms_received_flag_flag=0&sts_received_flag_flag=0&cmd=modem_main_state%2Cpin_status%2Cblc_wan_mode%2Cblc_wan_auto_mode%2Cloginfo%2Cfota_new_version_state%2Cfota_current_upgrade_state%2Cfota_upgrade_selector%2Cnetwork_provider%2Cis_mandatory%2Csta_count%2Cm_sta_count%2Csignalbar%2Cnetwork_type%2Csub_network_type%2Cppp_status%2Crj45_state%2CEX_SSID1%2Csta_ip_status%2CEX_wifi_profile%2Cm_ssid_enable%2Cwifi_cur_state%2CSSID1%2Csimcard_roam%2Clan_ipaddr%2Cbattery_charging%2Cbattery_vol_percent%2Cbattery_pers%2Cspn_name_data%2Cspn_b1_flag%2Cspn_b2_flag%2Crealtime_tx_bytes%2Crealtime_rx_bytes%2Crealtime_time%2Crealtime_tx_thrpt%2Crealtime_rx_thrpt%2Cmonthly_rx_bytes%2Cmonthly_tx_bytes%2Ctraffic_alined_delta%2Cmonthly_time%2Cdate_month%2Cdata_volume_limit_switch%2Cdata_volume_limit_size%2Cdata_volume_alert_percent%2Cdata_volume_limit_unit%2Croam_setting_option%2Cupg_roam_switch%2Cfota_package_already_download%2Cssid%2Cdial_mode%2Cethwan_mode%2Cdefault_wan_name%2Csms_received_flag%2Csts_received_flag%2Csms_unread_num"

const (
	PollInterval = 2 * time.Second
	cycleTimeout = 5 * time.Second
	maxBodyBytes = 1 << 20
)

type Client struct{ http *http.Client }

func NewClient() *Client {
	transport := &http.Transport{
		// Do not send LAN modem traffic through a system/environment proxy.
		Proxy:           nil,
		DialContext:     (&net.Dialer{Timeout: cycleTimeout, KeepAlive: 30 * time.Second}).DialContext,
		MaxConnsPerHost: 2, MaxIdleConns: 2, MaxIdleConnsPerHost: 2,
		IdleConnTimeout: 30 * time.Second, ResponseHeaderTimeout: cycleTimeout,
		MaxResponseHeaderBytes: 64 << 10,
	}
	return &Client{http: &http.Client{
		Transport: transport, Timeout: cycleTimeout,
		// A redirect may be a login page; do not follow it or leave the LAN host.
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

// Run owns one polling worker. Busy ticks are discarded rather than queued.
// Cancellation joins the in-flight cycle before returning to the window owner.
func (c *Client) Run(ctx context.Context, publish func(model.Update)) {
	defer c.http.CloseIdleConnections()
	if ctx.Err() != nil {
		return
	}
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()
	results := make(chan model.Update, 1)
	busy := false
	var finishedAt time.Time
	start := func() {
		busy = true
		go func() { results <- c.fetch(ctx) }()
	}
	start()
	for {
		select {
		case <-ctx.Done():
			if busy {
				<-results
			}
			return
		case tick := <-ticker.C:
			// A buffered tick dated before the last completion was also busy.
			if !busy && !tick.Before(finishedAt) && ctx.Err() == nil {
				start()
			}
		case update := <-results:
			busy = false
			finishedAt = time.Now()
			if ctx.Err() != nil {
				return
			}
			publish(update)
		}
	}
}

type failure struct {
	state   model.ConnectionState
	message string
}

func (c *Client) fetch(parent context.Context) model.Update {
	ctx, cancel := context.WithTimeout(parent, cycleTimeout)
	defer cancel()
	var radio signalResponse
	var status statusResponse
	results := make(chan failure, 2)
	go func() { results <- c.get(ctx, SignalURL, &radio, "Signal endpoint") }()
	go func() { results <- c.get(ctx, StatusURL, &status, "Status endpoint") }()
	first, second := <-results, <-results // Join both before reading their structs.
	// Prefer a concrete API failure over a concurrent transport failure.
	for _, result := range []failure{first, second} {
		if result.state == model.APIError {
			return model.Update{State: result.state, Message: result.message}
		}
	}
	for _, result := range []failure{first, second} {
		if result.state != "" {
			return model.Update{State: result.state, Message: result.message}
		}
	}
	return mapResponses(radio, status, time.Now())
}

func (c *Client) get(ctx context.Context, endpoint string, target any, label string) failure {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return failure{model.APIError, label + ": invalid request"}
	}
	req.Header.Set("Accept", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return failure{model.Disconnected, label + ": modem unreachable or request timed out"}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return failure{model.APIError, label + ": HTTP error or login required"}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes+1))
	if err != nil {
		return failure{model.Disconnected, label + ": response interrupted or timed out"}
	}
	if len(body) > maxBodyBytes {
		return failure{model.APIError, label + ": response exceeds 1 MiB"}
	}
	// The modem uses text/html even for JSON. Typed decoding ignores extra fields;
	// raw bodies and fields such as ICCID/SSID never enter a snapshot or a log.
	if err := json.Unmarshal(body, target); err != nil {
		return failure{model.APIError, label + ": invalid JSON or login page"}
	}
	return failure{}
}
