// Package config validates the single modem profile. Credentials must never be logged.
package config

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const DefaultHost = "http://192.168.100.1"
const DefaultPassword = "admin"

type ConnectionConfig struct {
	BaseURL  string
	Password string
}

func Validate(c ConnectionConfig) (ConnectionConfig, error) {
	host := strings.TrimSpace(c.BaseURL)
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	u, err := url.Parse(host)
	if err != nil || u == nil {
		return ConnectionConfig{}, errors.New("Host URL tidak valid")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil ||
		u.Opaque != "" || (u.Path != "" && u.Path != "/") || u.RawPath != "" || u.RawQuery != "" ||
		u.ForceQuery || u.Fragment != "" || strings.Contains(host, "#") {
		return ConnectionConfig{}, errors.New("Gunakan host HTTP/HTTPS tanpa path, query, fragment, atau username")
	}
	name := u.Hostname()
	if strings.Contains(name, ":") && !strings.HasPrefix(u.Host, "[") {
		return ConnectionConfig{}, errors.New("Alamat IPv6 harus memakai kurung siku")
	}
	if net.ParseIP(name) == nil {
		if len(name) > 253 {
			return ConnectionConfig{}, errors.New("Hostname terlalu panjang")
		}
		for _, label := range strings.Split(strings.TrimSuffix(name, "."), ".") {
			if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				return ConnectionConfig{}, errors.New("Hostname tidak valid")
			}
			for _, r := range label {
				if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
					return ConnectionConfig{}, errors.New("Hostname tidak valid")
				}
			}
		}
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return ConnectionConfig{}, errors.New("Port harus 1–65535")
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return ConnectionConfig{}, errors.New("Port tidak boleh kosong")
	}
	if c.Password == "" {
		return ConnectionConfig{}, errors.New("Password wajib diisi")
	}
	if len(c.Password) > 4096 {
		return ConnectionConfig{}, errors.New("Password terlalu panjang")
	}
	u.Host = strings.ToLower(u.Host)
	u.Path = ""
	c.BaseURL = u.String()
	return c, nil
}
