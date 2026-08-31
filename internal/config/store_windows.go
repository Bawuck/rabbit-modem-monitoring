package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

type diskProfile struct {
	Version  int    `json:"version"`
	Host     string `json:"host"`
	Password []byte `json:"passwordDPAPI"`
}

func profilePath() (string, error) {
	root := os.Getenv("LOCALAPPDATA")
	if root == "" || !filepath.IsAbs(root) {
		return "", errors.New("LOCALAPPDATA tidak tersedia")
	}
	return filepath.Join(root, "4G Monitor", "connection.json"), nil
}

func protect(data []byte, decrypt bool) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("Data kredensial kosong")
	}
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	var err error
	if decrypt {
		err = windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	} else {
		err = windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	}
	runtime.KeepAlive(data)
	if err != nil {
		return nil, errors.New("Windows DPAPI gagal memproses kredensial")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	buffer := unsafe.Slice(out.Data, int(out.Size))
	result := append([]byte(nil), buffer...)
	clear(buffer)
	return result, nil
}

func Load() (ConnectionConfig, error) {
	path, err := profilePath()
	if err != nil {
		return ConnectionConfig{}, err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return ConnectionConfig{}, nil
	}
	if err != nil {
		return ConnectionConfig{}, errors.New("Konfigurasi tidak dapat dibaca; isi ulang untuk mengganti")
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 64*1024+1))
	var saved diskProfile
	if err != nil || len(data) > 64*1024 || json.Unmarshal(data, &saved) != nil || saved.Version != 1 {
		return ConnectionConfig{}, errors.New("Konfigurasi rusak atau versi tidak didukung; isi ulang untuk mengganti")
	}
	password, err := protect(saved.Password, true)
	if err != nil {
		return ConnectionConfig{}, errors.New("Password tersimpan tidak dapat dibuka oleh pengguna Windows ini; isi ulang")
	}
	defer clear(password)
	c, err := Validate(ConnectionConfig{BaseURL: saved.Host, Password: string(password)})
	if err != nil {
		return ConnectionConfig{}, errors.New("Konfigurasi tersimpan tidak valid; isi ulang untuk mengganti")
	}
	return c, nil
}

func Save(c ConnectionConfig) error {
	c, err := Validate(c)
	if err != nil {
		return err
	}
	path, err := profilePath()
	if err != nil {
		return err
	}
	plain := []byte(c.Password)
	encrypted, err := protect(plain, false)
	clear(plain)
	if err != nil {
		return err
	}
	data, err := json.Marshal(diskProfile{Version: 1, Host: c.BaseURL, Password: encrypted})
	if err != nil {
		return errors.New("Konfigurasi tidak dapat disiapkan")
	}
	if os.MkdirAll(filepath.Dir(path), 0700) != nil {
		return errors.New("Folder konfigurasi tidak dapat dibuat")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".connection-*")
	if err != nil {
		return errors.New("File konfigurasi sementara tidak dapat dibuat")
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errors.New("Konfigurasi gagal ditulis; koneksi sebelumnya dipertahankan")
	}
	from, err := windows.UTF16PtrFromString(f.Name())
	if err != nil {
		return errors.New("Path konfigurasi tidak valid")
	}
	to, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return errors.New("Path konfigurasi tidak valid")
	}
	if windows.MoveFileEx(from, to, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH) != nil {
		return errors.New("Konfigurasi gagal diganti; koneksi sebelumnya dipertahankan")
	}
	return nil
}
