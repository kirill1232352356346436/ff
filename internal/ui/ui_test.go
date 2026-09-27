package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/kirill1232352356346436/go/internal/api"
	"github.com/kirill1232352356346436/go/internal/model"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestFiltersAndCheckoutValidation(t *testing.T) {
	products := demo().Products
	got := filterProducts(products, "", "Горный", "Сначала дешевле", true)
	if len(got) != 1 || got[0].Name != "Трейл 29" {
		t.Fatalf("unexpected filtered products: %+v", got)
	}
	products[0].Status = "rented"
	if len(filterProducts(products, "", "Все типы", "По порядку", true)) != 3 {
		t.Fatal("rented bike should be hidden")
	}
	got = filterProducts(products, "", "Все типы", "Сначала дешевле", false)
	if got[0].Name != "Джуниор 20" {
		t.Fatal("price sorting is incorrect")
	}
	if _, err := checkoutValues("Иван", "+79001234567", "2", "rent"); err != nil {
		t.Fatal(err)
	}
	if _, err := checkoutValues("Иван", "123", "2", "rent"); err == nil {
		t.Fatal("invalid phone was accepted")
	}
	if _, err := checkoutValues("Иван", "+79001234567", "0", "rent"); err == nil {
		t.Fatal("zero rental days were accepted")
	}
}

func demo() model.Snapshot {
	return model.Snapshot{Products: []model.Product{
		{ID: 1, Name: "Сити 28", Price: 28900, RentPerDay: 500, Category: "Городской", FrameSize: "M", Image: "city.png", Status: "available", Description: "Спокойные поездки по городу. Низкая рама, удобная посадка, крылья и багажник. Колёса 28 дюймов."},
		{ID: 2, Name: "Трейл 29", Price: 45900, RentPerDay: 850, Category: "Горный", FrameSize: "L", Image: "trail.png", Status: "available"},
		{ID: 3, Name: "Роуд 700", Price: 67900, RentPerDay: 1200, Category: "Шоссейный", FrameSize: "M", Image: "road.png", Status: "available"},
		{ID: 4, Name: "Джуниор 20", Price: 18900, RentPerDay: 350, Category: "Детский", FrameSize: "XS", Image: "junior.png", Status: "available"},
	}}
}

func TestNavigationAndCapture(t *testing.T) {
	application := test.NewApp()
	defer application.Quit()
	application.Settings().SetTheme(ShopTheme{})
	window := application.NewWindow("ВелоТочка")
	window.Resize(fyne.NewSize(1280, 1160))
	shop := New(window, api.New("http://127.0.0.1:8010"))
	shop.data = demo()
	shop.connection.Text = "Сервер подключён"
	shop.connection.Refresh()
	shop.render()
	window.Show()
	capture(t, window, "catalog.png")
	test.Tap(shop.nav["Заказы"])
	if shop.page != "Заказы" {
		t.Fatal("orders navigation failed")
	}
	test.Tap(shop.nav["Клиенты"])
	if shop.page != "Клиенты" {
		t.Fatal("customer navigation failed")
	}
	test.Tap(shop.nav["Каталог"])
	shop.checkout(shop.data.Products[0], "rent")
	capture(t, window, "rental.png")
	window.Close()
}

func capture(t *testing.T, window fyne.Window, name string) {
	directory := os.Getenv("VELO_SCREENSHOT_DIR")
	if directory == "" {
		return
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = png.Encode(file, window.Canvas().Capture()); err != nil {
		t.Fatal(err)
	}
}
