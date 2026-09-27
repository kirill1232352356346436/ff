package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/kirill1232352356346436/go/assets"
	"github.com/kirill1232352356346436/go/internal/api"
	"github.com/kirill1232352356346436/go/internal/ui"
	"os"
)

func main() {
	application := app.NewWithID("ru.velotochka.shop")
	application.Settings().SetTheme(ui.ShopTheme{})
	application.SetIcon(assets.Logo())
	window := application.NewWindow("ВелоТочка — магазин и прокат")
	window.Resize(fyne.NewSize(1200, 820))
	shop := ui.New(window, api.New(os.Getenv("VELO_API_URL")))
	shop.Reload()
	window.ShowAndRun()
}
