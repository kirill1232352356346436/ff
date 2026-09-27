// Package ui implements the new desktop interface. Widget state is owned by the UI thread.
package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"time"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/kirill1232352356346436/go/assets"
	"github.com/kirill1232352356346436/go/internal/api"
	"github.com/kirill1232352356346436/go/internal/model"
)

type Shop struct {
	window                     fyne.Window
	client                     *api.Client
	data                       model.Snapshot
	page                       string
	stage                      *fyne.Container
	notice                     *widget.Label
	connection                 *canvas.Text
	refreshButton              *widget.Button
	nav                        map[string]*widget.Button
	loading                    bool
	query, category, sortOrder string
	onlyAvailable              bool
}

func New(window fyne.Window, client *api.Client) *Shop {
	s := &Shop{window: window, client: client, page: "Каталог", category: "Все типы", sortOrder: "По порядку", nav: map[string]*widget.Button{}}
	s.stage = container.NewStack()
	s.notice = widget.NewLabel("")
	s.notice.Wrapping = fyne.TextWrapWord
	s.notice.Hide()
	s.connection = text("Подключение…", 12, pale, false)
	s.refreshButton = widget.NewButtonWithIcon("Обновить", theme.ViewRefreshIcon(), s.Reload)
	add := widget.NewButtonWithIcon("Добавить велосипед", theme.ContentAddIcon(), func() { s.editProduct(nil) })
	add.Importance = widget.HighImportance
	top := container.NewBorder(nil, nil, vbox(2, text("МАГАЗИН И ПРОКАТ", 11, muted, true), text("Каждая поездка начинается здесь", 20, ink, true)), hbox(8, s.refreshButton, add))
	main := container.NewBorder(vbox(10, top, s.notice), text("Учебный каталог · изображения моделей созданы для демонстрации", 11, muted, false), nil, nil, s.stage)
	window.SetPadded(false)
	window.SetContent(container.NewBorder(nil, nil, s.sidebar(), nil, inset(24, main)))
	s.render()
	return s
}

func (s *Shop) sidebar() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(assets.Logo())
	logo.SetMinSize(fyne.NewSize(42, 42))
	brand := vbox(14, hbox(10, logo, text("ВелоТочка", 23, white, true)), text("Продажа / аренда", 13, pale, false))
	nav := vbox(10)
	for _, tab := range []struct {
		name string
		icon fyne.Resource
	}{
		{"Каталог", theme.ListIcon()}, {"Заказы", theme.DocumentIcon()}, {"Клиенты", theme.AccountIcon()},
	} {
		name := tab.name
		button := widget.NewButtonWithIcon(name, tab.icon, func() { s.page = name; s.render() })
		button.Alignment = widget.ButtonAlignLeading
		s.nav[name] = button
		nav.Add(button)
	}
	top := vbox(32, brand, container.NewThemeOverride(nav, sideTheme{}))
	bottom := vbox(8, text("ЛОКАЛЬНЫЙ МАГАЗИН", 10, pale, true), s.connection)
	bg := canvas.NewRectangle(green)
	bg.SetMinSize(fyne.NewSize(220, 0))
	return container.NewStack(bg, inset(22, container.NewBorder(top, bottom, nil, nil)))
}

func (s *Shop) render() {
	for name, button := range s.nav {
		if name == s.page {
			button.Importance = widget.HighImportance
		} else {
			button.Importance = widget.LowImportance
		}
		button.Refresh()
	}
	var content fyne.CanvasObject
	switch s.page {
	case "Заказы":
		content = s.ordersPage()
	case "Клиенты":
		content = s.customersPage()
	default:
		content = s.catalogPage()
	}
	s.stage.Objects = []fyne.CanvasObject{content}
	s.stage.Refresh()
}

func (s *Shop) Reload() {
	if s.loading {
		return
	}
	s.loading = true
	s.refreshButton.Disable()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		data, err := s.client.Load(ctx)
		fyne.Do(func() {
			s.loading = false
			s.refreshButton.Enable()
			if err != nil {
				s.connection.Text = "Нет связи с сервером"
				s.connection.Refresh()
				s.message(err.Error(), true)
				return
			}
			s.data = data
			s.connection.Text = "Сервер подключён"
			s.connection.Refresh()
			if s.notice.Importance == widget.DangerImportance {
				s.notice.Hide()
			}
			s.render()
		})
	}()
}

func (s *Shop) message(message string, failed bool) {
	if failed {
		s.notice.Importance = widget.DangerImportance
	} else {
		s.notice.Importance = widget.SuccessImportance
	}
	s.notice.SetText(message)
	s.notice.Show()
}

func money(amount int) string {
	raw := fmt.Sprint(amount)
	for pos := len(raw) - 3; pos > 0; pos -= 3 {
		raw = raw[:pos] + " " + raw[pos:]
	}
	return raw + " ₽"
}
func shortDate(raw string) string {
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return raw
	}
	return parsed.Format("02.01.2006")
}

func emptyState(title, help string) fyne.CanvasObject {
	label := widget.NewLabel(help)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(vbox(10, text(title, 22, ink, true), label))
}
