package tests

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestDeletePickingTimesByID(t *testing.T) {
	t.Run("Получение заказа", func(t *testing.T) {
		err := TryRequest(t, "GET", "orders-service/1", nil, nil, nil)
		assert.Equal(t, http.StatusNotFound, err.(UnsuccessfulResponse).StatusCode)
	})
}

