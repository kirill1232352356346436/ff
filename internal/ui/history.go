package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/kirill1232352356346436/go/internal/model"
	"strings"
)

func (s *Shop) ordersPage() fyne.CanvasObject {
	list := vbox(12)
	filter := widget.NewSelect([]string{"Все заказы", "Действующая аренда", "Продажи", "Возвращённые"}, nil)
	filter.SetSelected("Все заказы")
	search := widget.NewEntry()
	search.SetPlaceHolder("Велосипед, клиент или телефон")
	update := func() {
		list.RemoveAll()
		for _, order := range s.data.Orders {
			active := order.Kind == "rent" && order.ReturnedDate == nil
			if filter.Selected == "Действующая аренда" && !active {
				continue
			}
			if filter.Selected == "Продажи" && order.Kind != "sale" {
				continue
			}
			if filter.Selected == "Возвращённые" && order.ReturnedDate == nil {
				continue
			}
			if !strings.Contains(strings.ToLower(order.ProductName+" "+order.CustomerName+" "+order.Phone), strings.ToLower(search.Text)) {
				continue
			}
			list.Add(s.orderCard(order))
		}
		if len(list.Objects) == 0 {
			list.Add(emptyState("Заказов пока нет", "Оформите продажу или аренду в каталоге."))
		}
	}
	filter.OnChanged = func(string) { update() }
	search.OnChanged = func(string) { update() }
	update()
	top := vbox(12, text("Заказы и аренда", 29, ink, true), text("История операций, сроки и возврат велосипедов.", 14, muted, false), container.NewBorder(nil, nil, nil, filter, search))
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(list))
}

func (s *Shop) orderCard(order model.Order) fyne.CanvasObject {
	kind := "Продажа"
	state := "Завершена"
	dates := "Оформлено " + shortDate(order.StartDate)
	if order.Kind == "rent" {
		kind = "Аренда"
		state = "На руках"
		if order.DueDate != nil {
			dates += " · вернуть до " + shortDate(*order.DueDate)
		}
		if order.Overdue {
			state = "Просрочено"
		}
		if order.ReturnedDate != nil {
			state = "Возвращён"
			dates += " · возврат " + shortDate(*order.ReturnedDate)
		}
	}
	line := widget.NewLabel(dates)
	line.Wrapping = fyne.TextWrapWord
	top := hbox(10, text(fmt.Sprintf("№%03d", order.ID), 14, muted, true), text(order.ProductName, 21, ink, true), layout.NewSpacer(), badge(kind+" · "+state, green))
	customer := widget.NewLabel(order.CustomerName + "  ·  " + order.Phone)
	bottom := hbox(10, text(money(order.Amount), 20, ink, true), layout.NewSpacer())
	if order.Kind == "rent" && order.ReturnedDate == nil {
		button := widget.NewButton("Принять возврат", nil)
		button.OnTapped = func() {
			dialog.ShowConfirm("Возврат велосипеда", "Подтвердить возврат «"+order.ProductName+"»?", func(ok bool) {
				if !ok {
					return
				}
				button.Disable()
				go func() {
					err := s.client.Return(context.Background(), order.ID)
					fyne.Do(func() {
						button.Enable()
						if err != nil {
							s.message(err.Error(), true)
							return
						}
						s.message("Велосипед возвращён в каталог", false)
						s.Reload()
					})
				}()
			}, s.window)
		}
		bottom.Add(button)
	}
	return card(vbox(5, top, customer, line, bottom))
}

func (s *Shop) customersPage() fyne.CanvasObject {
	list := vbox(12)
	search := widget.NewEntry()
	search.SetPlaceHolder("Поиск по имени или телефону")
	update := func() {
		list.RemoveAll()
		for _, customer := range s.data.Customers {
			if !strings.Contains(strings.ToLower(customer.Name+" "+customer.Phone), strings.ToLower(search.Text)) {
				continue
			}
			info := vbox(6, text(customer.Name, 21, ink, true), text(customer.Phone, 14, muted, false))
			amount := vbox(6, text(money(customer.TotalAmount), 20, green, true), text(fmt.Sprintf("Заказов: %d", customer.OrdersCount), 13, muted, false))
			list.Add(card(container.NewBorder(nil, nil, info, amount)))
		}
		if len(list.Objects) == 0 {
			list.Add(emptyState("Клиенты появятся после первого заказа", "Введите имя и телефон при продаже или аренде."))
		}
	}
	search.OnChanged = func(string) { update() }
	update()
	top := vbox(12, text("Клиенты", 29, ink, true), text("Контакты покупателей и арендаторов в одном месте.", 14, muted, false), search)
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(list))
}
