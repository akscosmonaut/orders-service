package main

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"http/models"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	pg "http/postgres"
)

var postgres pg.Connector

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, `{"statusCode": 400, "details": "EMPTY_BODY"}`, http.StatusBadRequest)
		log.Err(err).Msg("invalid body")
		return
	}

	var order models.BaseOrder
	err = json.Unmarshal(body, &order)
	if err != nil {
		http.Error(w, `{"statusCode": 400, "details": "INVALID_BODY"}`, http.StatusBadRequest)
		log.Err(err).Msg("invalid body")
		return
	}

	query, args, err := sqlx.In("INSERT INTO orders (name, address, phone, products, status, total) "+
		"(select ?, ?, ?, ?, ?, sum(price) as total from products where id = any (?)) "+
		"RETURNING id, status",
		order.Name, order.Address, order.Phone, pq.Array(order.Products), models.New, pq.Array(order.Products))
	query = postgres.DB.Rebind(query)
	var insertedID int64
	var insertedStatus string
	rows, err := postgres.DB.Queryx(query, args...)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("creating order")
		return
	}

	for rows.Next() {
		if err = rows.Scan(&insertedID, &insertedStatus); err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
			log.Err(err).Msg("creating order")
			return
		}
	}

	createdOrder := models.OrderCreated{
		ID:     insertedID,
		Status: insertedStatus,
	}

	result, err := json.Marshal(createdOrder)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("marshaling created order response")
		return
	}
	_, _ = w.Write(result)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		var order []models.FullOrder
		log.Info().Msg("no id is passed - get all orders")
		err := postgres.DB.Select(&order, "SELECT name, address, phone, products, status, total from orders")
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
			log.Err(err).Msg("cannot get orders")
			return
		}

		log.Info().Msgf("Got this orders %+v", &order)
		resp, _ := json.Marshal(&order)
		_, _ = w.Write(resp)
	} else {
		log.Info().Str("id", id).Msg("get one order")
		var order models.FullOrder
		err := postgres.DB.QueryRowx("SELECT name, address, phone, products, status, total from orders WHERE id = $1 LIMIT 1", id).
			StructScan(&order)
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
			log.Err(err).Str("id", id).Msg("cannot get order")
			return
		}

		log.Info().Msgf("Got this orders %+v", &order)
		resp, _ := json.Marshal(&order)
		_, _ = w.Write(resp)
	}

}

func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot update order status")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot update order status")
		return
	}

	var order models.BaseOrder
	err = json.Unmarshal(body, &order)
	_, err = postgres.DB.Exec("UPDATE orders SET name = $1, address = $2, phone = $3, products = $4 WHERE id=$5",
		order.Name, order.Address, order.Phone, pq.Array(order.Products), id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot update order status")
		return
	}

	_, _ = w.Write(nil)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_DELETE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot delete order")
		return
	}

	_, err := postgres.DB.Exec("DELETE FROM orders WHERE id=$1", id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_DELETE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot delete order")
		return
	}

	_, _ = w.Write(nil)
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Products
	err := postgres.DB.Select(&products, "SELECT * from products")
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_PRODUCTS"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot get products")
		return
	}

	log.Info().Msgf("Got this orders %+v", &products)
	resp, _ := json.Marshal(&products)
	_, _ = w.Write(resp)
}

func ChangeOrderStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot update order status")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot update order status")
		return
	}

	var st models.OrderStatus
	err = json.Unmarshal(body, &st)
	_, err = postgres.DB.Exec("UPDATE orders SET status = $1 WHERE id=$2", st.Status, id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot update order status")
		return
	}

	_, _ = w.Write(nil)
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	postgres.Connect()
	router := mux.NewRouter()
	router.HandleFunc("/orders-service/orders", CreateOrder).Methods("POST")
	router.HandleFunc("/orders-service/orders", GetOrders).Methods("GET")
	router.HandleFunc("/orders-service/orders/{id:[0-9]+}", GetOrders).Methods("GET")
	router.HandleFunc("/orders-service/orders/{id:[0-9]+}", UpdateOrder).Methods("PUT")
	router.HandleFunc("/orders-service/orders/{id:[0-9]+}", DeleteOrder).Methods("DELETE")
	router.HandleFunc("/orders-service/orders/{id:[0-9]+}/change-status", ChangeOrderStatus).Methods("PATCH")
	router.HandleFunc("/orders-service/products", GetProducts).Methods("GET")

	http.Handle("/", router)
	address := "0.0.0.0:9000"
	srv := &http.Server{
		Addr:         address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
	log.Info().Msgf("App starting... Trying: %s\n", address)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Err(err).Msg("problem with starting app")
		return
	}

}
