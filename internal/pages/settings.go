package pages

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"example.com/4g-monitor/internal/components"
	"example.com/4g-monitor/internal/config"
	"example.com/4g-monitor/internal/connection"
)

type Settings struct {
	host, password       widget.Editor
	toggle, save, cancel widget.Clickable
	list                 widget.List
	open, waiting        bool
	err                  string
	startup              widget.Bool
}

func (p *Settings) Open(view connection.View) {
	p.host = widget.Editor{SingleLine: true}
	p.password = widget.Editor{SingleLine: true, Mask: '•'}
	host := view.Config.BaseURL
	if host == "" {
		host = config.DefaultHost
	}
	p.host.SetText(host)
	password := view.Config.Password
	if !view.Ready && password == "" {
		password = config.DefaultPassword
	}
	p.password.SetText(password)
	p.list = widget.List{List: layout.List{Axis: layout.Vertical}}
	p.open, p.waiting, p.err = true, false, ""
	p.startup.Value = view.StartupEnabled
}

func (p *Settings) close() {
	p.password = widget.Editor{SingleLine: true, Mask: '•'}
	p.host.SetText("")
	p.open, p.waiting, p.err = false, false, ""
}

func (p *Settings) Layout(gtx layout.Context, t *material.Theme, controller *connection.Controller, view connection.View) layout.Dimensions {
	if !view.StartupBusy {
		p.startup.Value = view.StartupEnabled
		if !view.Busy && p.startup.Update(gtx) {
			controller.SetStartup(p.startup.Value)
		}
	}
	if p.waiting && !view.Busy {
		p.waiting = false
		if view.Error == "" {
			p.close()
			return layout.Dimensions{}
		}
	}
	if !view.Busy {
		// Consume pending edits before reading Text for a submit in this frame.
		for {
			if _, ok := p.host.Update(gtx); !ok {
				break
			}
		}
		for {
			if _, ok := p.password.Update(gtx); !ok {
				break
			}
		}
		for p.toggle.Clicked(gtx) {
			if p.password.Mask == 0 {
				p.password.Mask = '•'
			} else {
				p.password.Mask = 0
			}
		}
		for p.cancel.Clicked(gtx) {
			p.close()
			return layout.Dimensions{}
		}
		for p.save.Clicked(gtx) {
			value, err := config.Validate(config.ConnectionConfig{BaseURL: p.host.Text(), Password: p.password.Text()})
			if err != nil {
				p.err = err.Error()
			} else if controller.Submit(value) {
				p.err = ""
				p.waiting = true
				break
			}
		}
	}
	view = controller.Snapshot()
	if view.Busy {
		gtx = gtx.Disabled()
	}
	button := "Tampilkan password"
	if p.password.Mask == 0 {
		button = "Sembunyikan password"
	}
	message := p.err
	if message == "" {
		message = view.Error
	}
	if view.Busy {
		message = "Menyimpan dan menerapkan koneksi…"
	}
	rows := []layout.Widget{
		components.Label(t, 22, "Koneksi Modem", components.Foreground, true),
		components.Label(t, 12, "Host dan password disimpan untuk pengguna Windows ini. Password dilindungi Windows DPAPI.", components.Muted, false),
		components.Label(t, 12, "Host URL", components.Foreground, true),
		func(gtx layout.Context) layout.Dimensions {
			return material.Editor(t, &p.host, config.DefaultHost).Layout(gtx)
		},
		components.Label(t, 12, "Password modem", components.Foreground, true),
		func(gtx layout.Context) layout.Dimensions {
			return material.Editor(t, &p.password, "Password asli, bukan Base64").Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions { return material.Button(t, &p.toggle, button).Layout(gtx) },
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, components.Label(t, 14, "Start on startup", components.Foreground, true)),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if view.StartupBusy {
						gtx = gtx.Disabled()
					}
					return material.Switch(t, &p.startup, "Jalankan saat login Windows").Layout(gtx)
				}),
			)
		},
		components.Label(t, 11, "Langsung diterapkan untuk pengguna Windows ini, tanpa Simpan. EXE harus tetap di lokasi yang sama. Batal hanya membuang perubahan host/password.", components.Muted, false),
	}
	if view.StartupError != "" {
		rows = append(rows, components.Label(t, 12, view.StartupError, components.Amber, false))
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(p.host.Text())), "https://") {
		rows = append(rows, components.Label(t, 12, "HTTP mengirim password tanpa enkripsi jaringan. Gunakan hanya di LAN tepercaya.", components.Amber, false))
	}
	if message != "" {
		rows = append(rows, components.Label(t, 12, message, components.Amber, false))
	}
	rows = append(rows, func(gtx layout.Context) layout.Dimensions {
		return components.Row(gtx, 12,
			func(gtx layout.Context) layout.Dimensions {
				return material.Button(t, &p.save, "Simpan & Hubungkan").Layout(gtx)
			},
			func(gtx layout.Context) layout.Dimensions { return material.Button(t, &p.cancel, "Batal").Layout(gtx) },
		)
	})
	return layout.UniformInset(20).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.List(t, &p.list).Layout(gtx, len(rows), func(gtx layout.Context, i int) layout.Dimensions {
			return layout.Inset{Bottom: 16, Right: 8}.Layout(gtx, rows[i])
		})
	})
}
