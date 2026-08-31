package modem

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/4g-monitor/internal/model"
)

const loginPath = "/goform/goform_set_cmd_process"
const loginRetryInterval = time.Minute

func sessionActive(response monitorResponse) bool {
	return response.LoginInfo.valid && response.LoginInfo.isString && response.LoginInfo.text == "ok"
}

// Only the serial polling worker accesses login state. LOGIN is the only POST
// allowed here; no logout or device configuration commands are issued.
func (c *Client) login(ctx context.Context) failure {
	if c.password == "" {
		return failure{state: model.APIError, message: "Modem login required: buka Pengaturan dan isi password"}
	}
	if c.loginRejected {
		return failure{state: model.APIError, message: "Modem login rejected: periksa password di Pengaturan lalu Simpan & Hubungkan"}
	}
	if time.Now().Before(c.nextLogin) {
		return failure{state: model.APIError, message: "Modem login required; waiting before retry (60s limit)"}
	}
	if ctx.Err() != nil {
		return failure{state: model.Disconnected, message: "Modem login cancelled"}
	}
	c.nextLogin = time.Now().Add(loginRetryInterval)
	form := url.Values{
		"goformId": {"LOGIN"},
		"password": {base64.StdEncoding.EncodeToString([]byte(c.password))},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+loginPath, strings.NewReader(form.Encode()))
	if err != nil {
		return failure{state: model.APIError, message: "Invalid modem login request"}
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/index.html")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	response, err := c.http.Do(req)
	if err != nil {
		return failure{state: model.Disconnected, message: "Modem login unreachable or timed out"}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		c.loginRejected = true
	}
	if response.StatusCode != http.StatusOK {
		return failure{state: model.APIError, message: "Modem login HTTP error; redirect not followed"}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes+1))
	if err != nil {
		return failure{state: model.Disconnected, message: "Modem login response interrupted"}
	}
	var result struct {
		Result scalar `json:"result"`
	}
	if len(body) > maxBodyBytes || json.Unmarshal(body, &result) != nil || !result.Result.valid || result.Result.text == "" {
		return failure{state: model.APIError, message: "Invalid modem login response"}
	}
	if result.Result.text != "0" {
		// Do not keep submitting a rejected password and risk locking the modem.
		c.loginRejected = true
		return failure{state: model.APIError, message: "Modem login rejected: periksa password di Pengaturan lalu Simpan & Hubungkan"}
	}
	return failure{}
}
