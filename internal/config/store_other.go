//go:build !windows

package config

import "errors"

func Load() (ConnectionConfig, error) {
	return ConnectionConfig{}, errors.New("Penyimpanan koneksi memerlukan Windows DPAPI")
}
func Save(ConnectionConfig) error { return errors.New("Penyimpanan koneksi memerlukan Windows DPAPI") }
