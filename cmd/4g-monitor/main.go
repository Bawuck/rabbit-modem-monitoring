package main

import (
	"log"
	"os"

	"gioui.org/app"

	"example.com/4g-monitor/internal/windows"
)

func main() {
	go func() {
		if err := windows.Run(); err != nil {
			log.Print(err)
			os.Exit(1)
		}
		// Gio's platform loop owns the main thread. The widget owns the
		// application lifetime, including any secondary window.
		os.Exit(0)
	}()
	app.Main()
}
