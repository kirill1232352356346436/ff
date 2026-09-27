// Package api contains the HTTP client. No GUI code or polling channels live here.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kirill1232352356346436/go/internal/model"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New(base string) *Client {
	if strings.TrimSpace(base) == "" {
		base = "http://127.0.0.1:8010"
	}
	return &Client{BaseURL: strings.TrimRight(base, "/"), HTTP: &http.Client{Timeout: 8 * time.Second}}
}

func (c *Client) request(ctx context.Context, method, route string, input, output any) error {
	var body bytes.Buffer
	if input != nil {
		if err := json.NewEncoder(&body).Encode(input); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+route, &body)
	if err != nil {
		return fmt.Errorf("неверный адрес сервера: %w", err)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return errors.New("Нет связи с сервером. Запустите сервер и нажмите «Обновить».")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var payload struct {
			Detail json.RawMessage `json:"detail"`
		}
		if json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&payload) == nil {
			var detail string
			if json.Unmarshal(payload.Detail, &detail) == nil && detail != "" {
				return errors.New(detail)
			}
			var validation []struct {
				Msg string `json:"msg"`
				Loc []any  `json:"loc"`
			}
			if json.Unmarshal(payload.Detail, &validation) == nil && len(validation) > 0 {
				field := "Поле"
				if len(validation[0].Loc) > 0 {
					field = fmt.Sprint(validation[0].Loc[len(validation[0].Loc)-1])
				}
				names := map[string]string{"name": "Имя или название", "phone": "Телефон", "days": "Дни аренды", "price": "Цена", "rent_per_day": "Ставка аренды"}
				if translated, ok := names[field]; ok {
					field = translated
				}
				return fmt.Errorf("%s: проверьте введённое значение (%s)", field, validation[0].Msg)
			}
		}
		return fmt.Errorf("сервер вернул ошибку %d", resp.StatusCode)
	}
	if output == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("не удалось прочитать ответ сервера: %w", err)
	}
	return nil
}

func (c *Client) Load(ctx context.Context) (model.Snapshot, error) {
	var data model.Snapshot
	for _, item := range []struct {
		path   string
		output any
	}{
		{"/products", &data.Products}, {"/orders", &data.Orders}, {"/customers", &data.Customers},
	} {
		if err := c.request(ctx, "GET", item.path, nil, item.output); err != nil {
			return model.Snapshot{}, err
		}
	}
	return data, nil
}

func (c *Client) Checkout(ctx context.Context, id int, kind string, form model.Checkout) (model.Order, error) {
	var order model.Order
	if kind != "sale" && kind != "rent" {
		return order, errors.New("неизвестный вид заказа")
	}
	err := c.request(ctx, "POST", fmt.Sprintf("/products/%d/%s", id, kind), form, &order)
	return order, err
}

func (c *Client) Return(ctx context.Context, id int) error {
	return c.request(ctx, "POST", fmt.Sprintf("/orders/%d/return", id), nil, nil)
}

func (c *Client) SaveProduct(ctx context.Context, id int, form model.ProductInput) error {
	route, method := "/products", "POST"
	if id > 0 {
		route, method = fmt.Sprintf("/products/%d", id), "PUT"
	}
	return c.request(ctx, method, route, form, nil)
}

func (c *Client) DeleteProduct(ctx context.Context, id int) error {
	return c.request(ctx, "DELETE", fmt.Sprintf("/products/%d", id), nil, nil)
}
