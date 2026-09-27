// Package model defines the JSON contract with the Python API.
package model

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	RentPerDay  int    `json:"rent_per_day"`
	Category    string `json:"category"`
	FrameSize   string `json:"frame_size"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Status      string `json:"status"`
}

type ProductInput struct {
	Name        string `json:"name"`
	Price       int    `json:"price"`
	RentPerDay  int    `json:"rent_per_day"`
	Category    string `json:"category"`
	FrameSize   string `json:"frame_size"`
	Description string `json:"description"`
}

type Checkout struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Days  int    `json:"days"`
}

type Order struct {
	ID           int     `json:"id"`
	ProductID    int     `json:"product_id"`
	CustomerID   int     `json:"customer_id"`
	ProductName  string  `json:"product_name"`
	CustomerName string  `json:"customer_name"`
	Phone        string  `json:"phone"`
	Kind         string  `json:"kind"`
	StartDate    string  `json:"start_date"`
	DueDate      *string `json:"due_date"`
	ReturnedDate *string `json:"returned_date"`
	Amount       int     `json:"amount"`
	Overdue      bool    `json:"overdue"`
}

type Customer struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	OrdersCount int    `json:"orders_count"`
	TotalAmount int    `json:"total_amount"`
}

type Snapshot struct {
	Products  []Product
	Orders    []Order
	Customers []Customer
}
