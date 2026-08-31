// Package modem reads one fixed, local modem endpoint and renews its login.
// It never changes device settings or persists credentials/session data.
package modem

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"example.com/4g-monitor/internal/config"
	"example.com/4g-monitor/internal/model"
)

// Request only the fields consumed by the monitor in one modem response.
const monitorPath = "/goform/goform_get_cmd_process?multi_data=1&sms_received_flag_flag=0&sts_received_flag_flag=0&cmd=network_type%2Csub_network_type%2Crssi%2Clte_rsrp%2Cnv_rsrq%2Cnv_sinr%2Clte_band%2Cnv_pci%2Cppp_status%2Csignalbar%2Crealtime_rx_thrpt%2Crealtime_tx_thrpt%2Cnetwork_provider%2Csimcard_roam%2Csta_count%2Cm_sta_count%2Crealtime_rx_bytes%2Crealtime_tx_bytes%2Crealtime_time%2Cloginfo"

const (
	PollInterval = 2 * time.Second
	cycleTimeout = 5 * time.Second
	maxBodyBytes = 1 << 20
)

type ConnectionConfig = config.ConnectionConfig

type Client struct {
	baseURL       string
	monitorURL    string
	http          *http.Client
	password      string
	nextLogin     time.Time
	loginRejected bool
}

func NewClient(config ConnectionConfig) *Client {
	transport := &http.Transport{
		// Do not send LAN modem traffic through a system/environment proxy.
		Proxy:           nil,
		DialContext:     (&net.Dialer{Timeout: cycleTimeout, KeepAlive: 30 * time.Second}).DialContext,
		MaxConnsPerHost: 1, MaxIdleConns: 1, MaxIdleConnsPerHost: 1,
		IdleConnTimeout: 30 * time.Second, ResponseHeaderTimeout: cycleTimeout,
		MaxResponseHeaderBytes: 64 << 10,
	}
	jar, _ := cookiejar.New(nil) // No persistent cookie storage.
	return &Client{baseURL: config.BaseURL, monitorURL: config.BaseURL + monitorPath, password: config.Password, http: &http.Client{
		Jar:       jar,
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
	state         model.ConnectionState
	message       string
	loginRequired bool
}

func (c *Client) fetch(parent context.Context) model.Update {
	// Recovery can perform GET + LOGIN + GET; each request remains bounded to 5s.
	ctx, cancel := context.WithTimeout(parent, 3*cycleTimeout)
	defer cancel()
	var response monitorResponse
	result := c.get(ctx, c.monitorURL, &response, "Modem endpoint")
	if result.state != "" && !result.loginRequired {
		return model.Update{State: result.state, Message: result.message}
	}
	if result.loginRequired || !sessionActive(response) {
		if result = c.login(ctx); result.state != "" {
			return model.Update{State: result.state, Message: result.message}
		}
		response = monitorResponse{}
		if result = c.get(ctx, c.monitorURL, &response, "Modem endpoint"); result.state != "" {
			return model.Update{State: result.state, Message: result.message}
		}
		if !sessionActive(response) {
			return model.Update{State: model.APIError, Message: "Modem session unavailable after login; retry is rate-limited"}
		}
	}
	return mapResponse(response, time.Now())
}

func (c *Client) get(ctx context.Context, endpoint string, target any, label string) failure {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return failure{state: model.APIError, message: label + ": invalid request"}
	}
	req.Header.Set("Accept", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		return failure{state: model.Disconnected, message: label + ": modem unreachable or request timed out"}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return failure{state: model.APIError, message: label + ": HTTP error or login required",
			loginRequired: response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden ||
				(response.StatusCode >= 300 && response.StatusCode < 400)}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes+1))
	if err != nil {
		return failure{state: model.Disconnected, message: label + ": response interrupted or timed out"}
	}
	if len(body) > maxBodyBytes {
		return failure{state: model.APIError, message: label + ": response exceeds 1 MiB"}
	}
	// The modem uses text/html even for JSON. Typed decoding ignores extra fields;
	// raw bodies and fields such as ICCID/SSID never enter a snapshot or a log.
	if err := json.Unmarshal(body, target); err != nil {
		return failure{state: model.APIError, message: label + ": invalid JSON or login page"}
	}
	return failure{}
}
