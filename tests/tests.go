package tests

import (
	"net/http"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/akscosmonaut/orders-service"
)

func TestDeletePickingTimesByID(t *testing.T) {
	var resp map[string]interface{}


	t.Run("удаление несуществующего набор параметров", func(t *testing.T) {
		err := utils.TryRequest(t, "DELETE", "admin/v1/picking-times/9999", utils.AdminAuth(), nil, nil)
		assert.Equal(t, http.StatusNotFound, err.(utils.UnsuccessfulResponse).StatusCode)
	})

	//t.Run("удаление удаленного набор параметров", func(t *testing.T) {
	//	err := utils.TryRequest(t, "DELETE", "admin/v1/picking-times/553", utils.AdminAuth(), nil, nil)
	//	assert.Equal(t, http.StatusNotFound, err.(utils.UnsuccessfulResponse).StatusCode)
	//})
}

