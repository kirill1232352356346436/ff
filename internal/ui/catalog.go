package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kirill1232352356346436/go/assets"
	"github.com/kirill1232352356346436/go/internal/model"
	"sort"
	"strings"
)

func filterProducts(all []model.Product, query, category, order string, available bool) []model.Product {
	result := make([]model.Product, 0, len(all))
	for _, p := range all {
		if available && p.Status != "available" {
			continue
		}
		if category != "Все типы" && p.Category != category {
			continue
		}
		if !strings.Contains(strings.ToLower(p.Name+" "+p.Description), strings.ToLower(strings.TrimSpace(query))) {
			continue
		}
		result = append(result, p)
	}
	if order == "Сначала дешевле" {
		sort.SliceStable(result, func(i, j int) bool { return result[i].Price < result[j].Price })
	}
	if order == "Сначала дороже" {
		sort.SliceStable(result, func(i, j int) bool { return result[i].Price > result[j].Price })
	}
	return result
}

func (s *Shop) catalogPage() fyne.CanvasObject {
	count := map[string]int{}
	for _, p := range s.data.Products {
		count[p.Status]++
	}
	stat := func(title string, number int) fyne.CanvasObject {
		return card(hbox(14, text(fmt.Sprint(number), 27, green, true), text(title, 14, muted, false)))
	}
	stats := container.NewGridWithColumns(3, stat("В наличии", count["available"]), stat("В аренде", count["rented"]), stat("Продано", count["sold"]))
	heading := vbox(6, text("Выбери свой маршрут", 29, ink, true), text("Велосипеды для города, троп и новых открытий.", 14, muted, false))
	search := widget.NewEntry()
	search.SetPlaceHolder("Поиск велосипеда…")
	search.SetText(s.query)
	category := widget.NewSelect([]string{"Все типы", "Городской", "Горный", "Шоссейный", "Детский"}, nil)
	category.SetSelected(s.category)
	sorter := widget.NewSelect([]string{"По порядку", "Сначала дешевле", "Сначала дороже"}, nil)
	sorter.SetSelected(s.sortOrder)
	availability := widget.NewCheck("Только в наличии", nil)
	availability.SetChecked(s.onlyAvailable)
	cards := container.NewGridWithColumns(2)
	found := text("", 12, muted, false)
	area := container.NewStack(cards)
	update := func() {
		visible := filterProducts(s.data.Products, s.query, s.category, s.sortOrder, s.onlyAvailable)
		cards.RemoveAll()
		for _, bike := range visible {
			cards.Add(s.productCard(bike))
		}
		found.Text = fmt.Sprintf("Найдено: %d", len(visible))
		found.Refresh()
		if len(visible) == 0 {
			area.Objects = []fyne.CanvasObject{emptyState("Пока ничего не найдено", "Измените фильтры или добавьте велосипед.")}
		} else {
			area.Objects = []fyne.CanvasObject{cards}
		}
		area.Refresh()
	}
	search.OnChanged = func(value string) { s.query = value; update() }
	category.OnChanged = func(value string) { s.category = value; update() }
	sorter.OnChanged = func(value string) { s.sortOrder = value; update() }
	availability.OnChanged = func(value bool) { s.onlyAvailable = value; update() }
	update()
	filters := container.NewBorder(nil, nil, nil, hbox(8, category, sorter), search)
	top := vbox(16, heading, stats, filters, container.NewBorder(nil, nil, availability, found))
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(area))
}

func (s *Shop) productCard(p model.Product) fyne.CanvasObject {
	photo := canvas.NewImageFromResource(assets.Photo(p.Image))
	photo.FillMode = canvas.ImageFillContain
	photo.SetMinSize(fyne.NewSize(290, 150))
	photoBg := canvas.NewRectangle(paper)
	photoBg.CornerRadius = 10
	name := widget.NewLabelWithStyle(p.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	name.Truncation = fyne.TextTruncateEllipsis
	subtitle := text(p.Category+" · рама "+p.FrameSize, 12, muted, false)
	status := badge(statusLabel(p.Status), green)
	if p.Status == "rented" {
		status = badge(statusLabel(p.Status), amber)
	}
	edit := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() { s.editProduct(&p) })
	edit.Importance = widget.LowImportance
	top := container.NewBorder(nil, nil, subtitle, edit)
	title := container.NewBorder(nil, nil, nil, status, name)
	buy := widget.NewButton("Купить", func() { s.checkout(p, "sale") })
	buy.Importance = widget.HighImportance
	rent := widget.NewButton("В аренду", func() { s.checkout(p, "rent") })
	if p.Status != "available" {
		buy.Disable()
		rent.Disable()
	}
	pricing := hbox(12, text(money(p.Price), 23, ink, true), layout.NewSpacer(), text(money(p.RentPerDay)+" / день", 13, green, true))
	content := vbox(8, top, container.NewStack(photoBg, photo), title, pricing, container.NewGridWithColumns(2, buy, rent))
	return card(content)
}
