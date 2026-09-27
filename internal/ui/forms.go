package ui

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kirill1232352356346436/go/assets"
	"github.com/kirill1232352356346436/go/internal/model"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func checkoutValues(name, phone, days, kind string) (model.Checkout, error) {
	result := model.Checkout{Name: strings.TrimSpace(name), Phone: strings.TrimSpace(phone), Days: 1}
	if len([]rune(result.Name)) < 2 {
		return result, fmt.Errorf("введите имя клиента (от 2 символов)")
	}
	digits := 0
	for _, r := range result.Phone {
		if r >= '0' && r <= '9' {
			digits++
			continue
		}
		if !unicode.IsSpace(r) && !strings.ContainsRune("+()-", r) {
			return result, fmt.Errorf("проверьте номер телефона")
		}
	}
	if digits < 7 || digits > 15 {
		return result, fmt.Errorf("в телефоне должно быть от 7 до 15 цифр")
	}
	if kind == "rent" {
		n, err := strconv.Atoi(strings.TrimSpace(days))
		if err != nil || n < 1 || n > 365 {
			return result, fmt.Errorf("аренда оформляется на срок от 1 до 365 дней")
		}
		result.Days = n
	}
	return result, nil
}

func (s *Shop) checkout(product model.Product, kind string) {
	name, phone, days := widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	name.SetPlaceHolder("Имя и фамилия")
	phone.SetPlaceHolder("+7 900 123-45-67")
	days.SetText("1")
	errorLabel := widget.NewLabel("")
	errorLabel.Importance = widget.DangerImportance
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.Hide()
	total := text("", 24, green, true)
	due := widget.NewLabel("")
	description := widget.NewLabel(product.Description)
	description.Wrapping = fyne.TextWrapWord
	photo := canvas.NewImageFromResource(assets.Photo(product.Image))
	photo.FillMode = canvas.ImageFillContain
	photo.SetMinSize(fyne.NewSize(360, 170))
	form := widget.NewForm(widget.NewFormItem("Клиент", name), widget.NewFormItem("Телефон", phone))
	title, action := "Покупка велосипеда", "Оформить продажу"
	if kind == "rent" {
		title, action = "Аренда велосипеда", "Оформить аренду"
		form.Append("Дней аренды", days)
		due.SetText("Стоимость: " + money(product.RentPerDay) + " за день")
	}
	updateTotal := func(string) {
		amount := product.Price
		if kind == "rent" {
			n, err := strconv.Atoi(strings.TrimSpace(days.Text))
			if err != nil || n < 1 || n > 365 {
				total.Text = "Проверьте срок аренды"
				total.Refresh()
				return
			}
			amount = n * product.RentPerDay
			due.SetText("Возврат до " + time.Now().AddDate(0, 0, n).Format("02.01.2006") + " · " + money(product.RentPerDay) + " / день")
		}
		total.Text = "Итого: " + money(amount)
		total.Refresh()
	}
	days.OnChanged = updateTotal
	updateTotal("")
	confirm := widget.NewButton(action, nil)
	confirm.Importance = widget.HighImportance
	cancel := widget.NewButton("Отмена", nil)
	body := vbox(10, photo, text(product.Name, 24, ink, true), description, form, due, total, errorLabel, container.NewGridWithColumns(2, cancel, confirm))
	modal := dialog.NewCustomWithoutButtons(title, body, s.window)
	modal.Resize(fyne.NewSize(510, 670))
	cancel.OnTapped = modal.Hide
	confirm.OnTapped = func() {
		input, err := checkoutValues(name.Text, phone.Text, days.Text, kind)
		if err != nil {
			errorLabel.SetText(err.Error())
			errorLabel.Show()
			return
		}
		confirm.Disable()
		cancel.Disable()
		name.Disable()
		phone.Disable()
		days.Disable()
		errorLabel.Hide()
		go func() {
			order, err := s.client.Checkout(context.Background(), product.ID, kind, input)
			fyne.Do(func() {
				confirm.Enable()
				cancel.Enable()
				name.Enable()
				phone.Enable()
				days.Enable()
				if err != nil {
					errorLabel.SetText(err.Error())
					errorLabel.Show()
					return
				}
				modal.Hide()
				s.message(fmt.Sprintf("Заказ №%03d оформлен · %s · %s", order.ID, product.Name, money(order.Amount)), false)
				s.Reload()
			})
		}()
	}
	modal.Show()
}

func (s *Shop) editProduct(existing *model.Product) {
	name, price, rate, frame := widget.NewEntry(), widget.NewEntry(), widget.NewEntry(), widget.NewEntry()
	description := widget.NewMultiLineEntry()
	description.SetMinRowsVisible(2)
	category := widget.NewSelect([]string{"Городской", "Горный", "Шоссейный", "Детский"}, nil)
	category.SetSelected("Городской")
	frame.SetText("M")
	photoKeys := map[string]string{"Городской": "city.png", "Горный": "trail.png", "Шоссейный": "road.png", "Детский": "junior.png"}
	photo := canvas.NewImageFromResource(assets.Photo("city.png"))
	photo.FillMode = canvas.ImageFillContain
	photo.SetMinSize(fyne.NewSize(350, 130))
	category.OnChanged = func(value string) { photo.Resource = assets.Photo(photoKeys[value]); photo.Refresh() }
	title := "Новый велосипед"
	id := 0
	if existing != nil {
		title = "Изменить велосипед"
		id = existing.ID
		name.SetText(existing.Name)
		price.SetText(strconv.Itoa(existing.Price))
		rate.SetText(strconv.Itoa(existing.RentPerDay))
		frame.SetText(existing.FrameSize)
		category.SetSelected(existing.Category)
		description.SetText(existing.Description)
	}
	form := widget.NewForm(widget.NewFormItem("Название", name), widget.NewFormItem("Тип", category), widget.NewFormItem("Рама", frame), widget.NewFormItem("Цена, ₽", price), widget.NewFormItem("Аренда, ₽/день", rate), widget.NewFormItem("Описание", description))
	errorLabel := widget.NewLabel("")
	errorLabel.Wrapping = fyne.TextWrapWord
	errorLabel.Importance = widget.DangerImportance
	errorLabel.Hide()
	save := widget.NewButton("Сохранить", nil)
	save.Importance = widget.HighImportance
	cancel := widget.NewButton("Отмена", nil)
	actions := container.NewGridWithColumns(2, cancel, save)
	body := vbox(8, photo, form, errorLabel, actions)
	modal := dialog.NewCustomWithoutButtons(title, body, s.window)
	modal.Resize(fyne.NewSize(520, 690))
	cancel.OnTapped = modal.Hide
	save.OnTapped = func() {
		p, err1 := strconv.Atoi(strings.TrimSpace(price.Text))
		r, err2 := strconv.Atoi(strings.TrimSpace(rate.Text))
		if len([]rune(strings.TrimSpace(name.Text))) < 2 || strings.TrimSpace(frame.Text) == "" || err1 != nil || err2 != nil || p < 1 || r < 1 {
			errorLabel.SetText("Введите название, размер рамы и положительные целые цены.")
			errorLabel.Show()
			return
		}
		input := model.ProductInput{Name: strings.TrimSpace(name.Text), Price: p, RentPerDay: r, Category: category.Selected, FrameSize: strings.TrimSpace(frame.Text), Description: strings.TrimSpace(description.Text)}
		save.Disable()
		cancel.Disable()
		go func() {
			err := s.client.SaveProduct(context.Background(), id, input)
			fyne.Do(func() {
				save.Enable()
				cancel.Enable()
				if err != nil {
					errorLabel.SetText(err.Error())
					errorLabel.Show()
					return
				}
				modal.Hide()
				s.message("Велосипед сохранён", false)
				s.Reload()
			})
		}()
	}
	if existing != nil {
		remove := widget.NewButton("Удалить из каталога", func() {
			dialog.ShowConfirm("Удаление", "Удалить «"+existing.Name+"»? Велосипед с историей заказов удалить нельзя.", func(ok bool) {
				if !ok {
					return
				}
				save.Disable()
				cancel.Disable()
				go func() {
					err := s.client.DeleteProduct(context.Background(), existing.ID)
					fyne.Do(func() {
						save.Enable()
						cancel.Enable()
						if err != nil {
							errorLabel.SetText(err.Error())
							errorLabel.Show()
							return
						}
						modal.Hide()
						s.message("Велосипед удалён", false)
						s.Reload()
					})
				}()
			}, s.window)
		})
		remove.Importance = widget.LowImportance
		body.Add(remove)
	}
	modal.Show()
}
