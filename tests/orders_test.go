package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestOrdersByID(t *testing.T) {

	t.Run("Создание заказа", func(t *testing.T) {
		req := json.RawMessage(`{
			"address":"test address",
			"name": "Regina",
			"phone":"+79267484433",
			"products":[1,2,6,7]
		}`)
		Request(t, "POST", "orders-service/orders", nil, req, nil)
	})

	t.Run("Получение заказа", func(t *testing.T) {
		Request(t, "GET", "orders-service/orders/1", nil, nil, nil)
	})

	t.Run("Получение несуществующего заказа", func(t *testing.T) {
		err := TryRequest(t, "GET", "orders-service/orders/200", nil, nil, nil)
		assert.Equal(t, http.StatusNotFound, err.(UnsuccessfulResponse).StatusCode)
	})

	t.Run("Обновление заказа", func(t *testing.T) {
		req := json.RawMessage(`{
			"address":"test address",
			"name": "Maxim",
			"phone":"+79267484455",
			"products":[1,2,6,7]
		}`)
		Request(t, "PUT", "orders-service/orders/1", nil, req, nil)
	})

	t.Run("Удаление заказа", func(t *testing.T) {
		Request(t, "DELETE", "orders-service/orders/1", nil, nil, nil)
	})

	t.Run("Обновление статуса заказа", func(t *testing.T) {
		req := json.RawMessage(`{
			"status":"canceled"
		}`)
		Request(t, "PATCH", "orders-service/orders/3/change-status", nil, req, nil)
	})
}
