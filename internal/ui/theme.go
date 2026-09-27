package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

var (
	ink   = color.NRGBA{R: 25, G: 48, B: 42, A: 255}
	muted = color.NRGBA{R: 110, G: 125, B: 118, A: 255}
	green = color.NRGBA{R: 23, G: 94, B: 75, A: 255}
	pale  = color.NRGBA{R: 228, G: 242, B: 226, A: 255}
	paper = color.NRGBA{R: 247, G: 248, B: 244, A: 255}
	line  = color.NRGBA{R: 223, G: 230, B: 223, A: 255}
	white = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	amber = color.NRGBA{R: 163, G: 104, B: 32, A: 255}
)

type ShopTheme struct{}

func (ShopTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return paper
	case theme.ColorNameForeground:
		return ink
	case theme.ColorNamePrimary, theme.ColorNameHyperlink:
		return green
	case theme.ColorNameButton:
		return pale
	case theme.ColorNameInputBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return white
	case theme.ColorNameInputBorder, theme.ColorNameSeparator:
		return line
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return muted
	case theme.ColorNameDisabledButton:
		return line
	case theme.ColorNameHover, theme.ColorNamePressed:
		return color.NRGBA{R: 23, G: 94, B: 75, A: 30}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 23, G: 94, B: 75, A: 70}
	case theme.ColorNameForegroundOnPrimary:
		return white
	}
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}
func (ShopTheme) Font(style fyne.TextStyle) fyne.Resource    { return theme.DefaultTheme().Font(style) }
func (ShopTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(name) }
func (ShopTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameInputRadius, theme.SizeNameButtonRadius:
		return 8
	case theme.SizeNameCardRadius, theme.SizeNameDialogRadius:
		return 14
	}
	return theme.DefaultTheme().Size(name)
}

type sideTheme struct{ ShopTheme }

func (sideTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground:
		return white
	case theme.ColorNameButton:
		return color.NRGBA{R: 38, G: 85, B: 68, A: 255}
	case theme.ColorNamePrimary:
		return pale
	case theme.ColorNameForegroundOnPrimary:
		return green
	case theme.ColorNameHover, theme.ColorNamePressed:
		return color.NRGBA{R: 220, G: 245, B: 225, A: 30}
	}
	return ShopTheme{}.Color(name, variant)
}

func text(s string, size float32, c color.Color, bold bool) *canvas.Text {
	t := canvas.NewText(s, c)
	t.TextSize, t.TextStyle = size, fyne.TextStyle{Bold: bold}
	return t
}
func inset(amount float32, obj fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewCustomPaddedLayout(amount, amount, amount, amount), obj)
}
func card(obj fyne.CanvasObject) *fyne.Container {
	bg := canvas.NewRectangle(white)
	bg.CornerRadius = 14
	bg.StrokeColor = line
	bg.StrokeWidth = 1
	return container.NewStack(bg, inset(16, obj))
}
func hbox(gap float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewCustomPaddedHBoxLayout(gap), objects...)
}
func vbox(gap float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(layout.NewCustomPaddedVBoxLayout(gap), objects...)
}
func badge(label string, tint color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(pale)
	bg.CornerRadius = 6
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(4, 4, 8, 8), text(label, 11, tint, true)))
}
func statusLabel(status string) string {
	switch status {
	case "available":
		return "В наличии"
	case "rented":
		return "В аренде"
	case "sold":
		return "Продан"
	}
	return status
}
