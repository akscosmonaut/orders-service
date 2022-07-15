package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Person struct {
	Name     string  `db:"name" json:"name"`
	Address  string  `db:"address" json:"address"`
	Phone    string  `db:"phone" json:"phone"`
}
/*
type Product struct {
	ID int `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Price int `db:"price" json:"price"`
}*/

type OrderStatus struct {
	Status string `db:"status" json:"status"`
}

type Order struct {
	ID int `db:"id" json:"id"`
	Person
	Products `db:"products" json:"products"`
	OrderStatus
	Total int `db:"total" json:"total"`
}

type CreateOrder struct {
	Person
	Products []int `db:"products" json:"products"`
}

type OrderCreated struct {
	ID     int64  `db:"id" json:"id"`
	Status string `db:"status" json:"status"`
}

type Product struct {
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

type Products []Product

func (pr Products) Scan(value interface{}) error {
	var b []byte
	switch t := value.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	default:
		return errors.New("unknown type")
	}

	err := json.Unmarshal(b, &pr)
	if err != nil {
		return errors.New("p")
	}
	return nil
}

func (pr Products) Value() (driver.Value, error) {
	b, err := json.Marshal(&pr)
	if err != nil {
		return nil, errors.New("dd")
	}
	return b, nil
}

func (pr *Product) Scan(value interface{}) error {
	var b []byte
	switch t := value.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	default:
		return errors.New("unknown type")
	}

	err := json.Unmarshal(b, &pr)
	if err != nil {
		return errors.New("p")
	}
	return nil
}

func (pr Product) Value() (driver.Value, error) {
	b, err := json.Marshal(&pr)
	if err != nil {
		return nil, errors.New("dd")
	}
	return b, nil
}