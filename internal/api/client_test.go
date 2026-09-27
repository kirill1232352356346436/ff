package api

import (
	"context"
	"github.com/kirill1232352356346436/go/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckoutSendsJSONAndReadsNullDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/products/42/sale" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing JSON content type")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":7,"product_id":42,"kind":"sale","amount":25000,"due_date":null,"returned_date":null}`))
	}))
	defer server.Close()
	order, err := New(server.URL).Checkout(context.Background(), 42, "sale", model.Checkout{Name: "Иван", Phone: "+79991234567", Days: 1})
	if err != nil || order.ID != 7 || order.DueDate != nil {
		t.Fatalf("order=%+v error=%v", order, err)
	}
}

func TestServerConflictIsShownToUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`{"detail":"Велосипед уже продан"}`))
	}))
	defer server.Close()
	_, err := New(server.URL).Checkout(context.Background(), 1, "rent", model.Checkout{})
	if err == nil || !strings.Contains(err.Error(), "Велосипед уже продан") {
		t.Fatalf("error=%v", err)
	}
}
