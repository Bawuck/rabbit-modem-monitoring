//go:build !windows

package startup

import "errors"

func Status() (bool, error) { return false, errors.New("Startup otomatis hanya tersedia di Windows") }
func Set(bool) error        { return errors.New("Startup otomatis hanya tersedia di Windows") }
