// Package components contains reusable Gio-only presentation components.
package components

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

var (
	Background = rgb(0x0c121c)
	Surface    = rgb(0x151f2e)
	Border     = rgb(0x2a394c)
	Foreground = rgb(0xeaf1fa)
	Muted      = rgb(0x9aacc3)
	Accent     = rgb(0x56d9d0)
	Green      = rgb(0x7ee0ad)
	Amber      = rgb(0xf7c873)
	Red        = rgb(0xff8994)
	Blue       = rgb(0x87b7ff)
)

func rgb(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

// Each window needs its own theme and text shaper.
func NewTheme() *material.Theme {
	t := material.NewTheme()
	t.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))
	t.Palette = material.Palette{Bg: Background, Fg: Foreground, ContrastBg: Accent, ContrastFg: Background}
	t.TextSize = 14
	return t
}

func Label(t *material.Theme, size unit.Sp, value string, col color.NRGBA, bold bool) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Label(t, size, value)
		l.Color = col
		if bold {
			l.Font.Weight = font.Bold
		}
		return l.Layout(gtx)
	}
}

func Column(gtx layout.Context, gap unit.Dp, children ...layout.Widget) layout.Dimensions {
	items := make([]layout.FlexChild, 0, len(children)*2)
	for i, child := range children {
		if i > 0 {
			items = append(items, layout.Rigid(layout.Spacer{Height: gap}.Layout))
		}
		items = append(items, layout.Rigid(child))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, items...)
}

// Row shares the available width evenly, with a fixed gap between columns.
func Row(gtx layout.Context, gap unit.Dp, children ...layout.Widget) layout.Dimensions {
	items := make([]layout.FlexChild, 0, len(children)*2)
	for i, child := range children {
		if i > 0 {
			items = append(items, layout.Rigid(layout.Spacer{Width: gap}.Layout))
		}
		items = append(items, layout.Flexed(1, child))
	}
	return layout.Flex{Alignment: layout.Start}.Layout(gtx, items...)
}

func Card(gtx layout.Context, content layout.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, Surface, clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(12)).Op(gtx.Ops))
			return layout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(12).Layout(gtx, content)
		},
	)
}

func Badge(gtx layout.Context, t *material.Theme, value string, col color.NRGBA) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			paint.FillShape(gtx.Ops, Border, clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(6)).Op(gtx.Ops))
			return layout.Dimensions{Size: gtx.Constraints.Min}
		},
		func(gtx layout.Context) layout.Dimensions {
			return (layout.Inset{Top: 3, Bottom: 3, Left: 7, Right: 7}).Layout(gtx, Label(t, 11, value, col, true))
		},
	)
}
