package models

type BaseOrder struct {
	Name     string  `db:"name" json:"name"`
	Address  string  `db:"address" json:"address"`
	Phone    string  `db:"phone" json:"phone"`
	Products []uint8 `db:"products" json:"products"`
}

type OrderStatus struct {
	Status string `db:"status" json:"status"`
}

type FullOrder struct {
	BaseOrder
	OrderStatus
	Total int `db:"total" json:"total"`
}

type OrderCreated struct {
	ID     int64  `db:"id" json:"id"`
	Status string `db:"status" json:"status"`
}

type Products struct {
	ID    int64   `db:"id" json:"id"`
	Name  string  `db:"name" json:"name"`
	Price float64 `db:"price" json:"price"`
}

const (
	New       = "new"
	Confirmed = "confirmed"
	Canceled  = "canceled"
	Done      = "done"
)
