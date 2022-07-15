package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	var order models.CreateOrder
	err = json.Unmarshal(body, &order)
	if err != nil {
		http.Error(w, `{"statusCode": 400, "details": "INVALID_BODY"}`, http.StatusBadRequest)
		log.Err(err).Msg("invalid body")
		return
	}

	// open transaction
	tx, err := postgres.DB.Beginx()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("creating order")
		return
	}
	// rollback transaction
	defer tx.Rollback()

	query, args, err := sqlx.In(`INSERT INTO orders (name, address, phone, status, total)
		(select ?, ?, ?, ?, sum(price) as total from products where id = any (?))
		RETURNING id, status`,
		order.Name, order.Address, order.Phone, models.New, pq.Array(order.Products))
	query = tx.Rebind(query)
	var insertedID int64
	var insertedStatus string
	rows, err := tx.Queryx(query, args...)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("inserting order into orders")
		return
	}

	for rows.Next() {
		if err = rows.Scan(&insertedID, &insertedStatus); err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
			log.Err(err).Msg("scan insert order result")
			return
		}
	}

	createdOrder := models.OrderCreated{
		ID:     insertedID,
		Status: insertedStatus,
	}

	stmt, err := tx.Preparex("insert into orders_products (order_id, product_id) VALUES ($1, $2);")
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("preparing order product transaction")
		return
	}
	defer stmt.Close()

	// insert into order_products
	for product := range order.Products {
		_, err = stmt.Exec(createdOrder.ID, product)
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
			log.Err(err).Msg("exec order product transaction")
			return
		}
	}
	// commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("commit order product transaction")
		return
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
		var orders []models.Order
		log.Info().Msg("no id is passed - get all orders")
		rows, err := postgres.DB.Queryx( `
				SELECT ord.id as id , ord.name as name, address, phone, status, total,
       				array_agg(json_build_object('id', p.id, 'name', p.name, 'price', p.price)) as products
				from orders as ord
         			left join orders_products op on ord.id = op.order_id
         			left join products p on op.product_id = p.id
				group by ord.id;`)
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
			log.Err(err).Msg("cannot get ordersss")
			return
		}
		for rows.Next() {
			var order models.Order
			if err := rows.Scan(&order.ID, &order.Name, &order.Address, &order.Phone, &order.Status,
				&order.Total, pq.Array(&order.Products)); err != nil {
				http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
				log.Err(err).Msg("cannot get orders")
				return
			}
			orders = append(orders, order)
		}
		log.Info().Msgf("Got this orders %+v", &orders)
		resp, _ := json.Marshal(&orders)
		_, _ = w.Write(resp)
	} else {
		log.Info().Str("id", id).Msg("get one order")
		rows, err := postgres.DB.Queryx(`SELECT ord.id as id, ord.name as name, address, phone, status, total,
			array_agg(json_build_object('id', p.id, 'name', p.name, 'price', p.price)) as products 
			from orders as ord 
			left join orders_products op on ord.id = op.order_id 
			left join products p on op.product_id = p.id  WHERE ord.id = $1 group by ord.id LIMIT 1;`, id)
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
			log.Err(err).Str("id", id).Msg("cannot get order")
			return
		}

		var order models.Order
		for rows.Next() {
			if err := rows.Scan(&order.ID, &order.Name, &order.Address, &order.Phone, &order.Status,
				&order.Total, pq.Array(&order.Products)); err != nil {
				http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_ORDERS"}`, http.StatusInternalServerError)
				log.Err(err).Msg("cannot get ordersss")
				return
			}
		}

		if order.ID == 0 && len(order.Products) == 0 {
			http.Error(w, `{"statusCode": 404, "details": "ORDER_NOT_FOUND"}`, http.StatusNotFound)
			log.Err(err).Msg("order not found")
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

	var order models.CreateOrder
	err = json.Unmarshal(body, &order)

	tx, err := postgres.DB.Beginx()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("open transaction for order update")
		return
	}
	// rollback transaction
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE orders SET name = $1, address = $2, phone = $3 WHERE id=$4",
		order.Name, order.Address, order.Phone, id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot update order status")
		return
	}

	n, _ := res.RowsAffected()
	if  n == 0 {
		http.Error(w, `{"statusCode": 404, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusNotFound)
		log.Err(errors.New("order not found")).Msg("cannot update order status")
		return
	}

	_, err = tx.Exec("DELETE FROM orders_products WHERE order_id=$1", id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot update order status")
		return
	}

	stmt, err := tx.Preparex("insert into orders_products (order_id, product_id) VALUES ($1, $2);")
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("preparing order product transaction")
		return
	}
	defer stmt.Close()

	// insert into order_products
	for _, product := range order.Products {
		_, err = stmt.Exec(id, product)
		if err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
			log.Err(err).Msg("exec order product transaction")
			return
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_CREATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("commit order product transaction")
		return
	}

	resp := fmt.Sprintf(`{"order_id": "%s", "details": "updated"}`, id)
	_, _ = w.Write([]byte(resp))
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		http.Error(w, `{"statusCode": 400, "details": "CANNOT_DELETE_ORDER"}`, http.StatusBadRequest)
		log.Error().Msg("cannot delete order")
		return
	}

	tx, err := postgres.DB.Beginx()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_DELETE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("prepare transaction for deleting order")
		return
	}
	// rollback transaction
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM orders WHERE id=$1", id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_DELETE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot delete order from orders table")
		return
	}

	_, err = tx.Exec("DELETE FROM orders_products WHERE order_id=$1", id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_DELETE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot delete order from orders_products table")
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_DELETE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot commit transaction with delete order")
		return
	}

	resp := fmt.Sprintf(`{"order_id": "%s", "details": "deleted"}`, id)
	_, _ = w.Write([]byte(resp))
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := postgres.DB.Queryx("SELECT id, name, price from products")
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_PRODUCTS"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot get products")
		return
	}

	var products models.Products
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			http.Error(w, `{"statusCode": 500, "details": "CANNOT_GET_PRODUCTS"}`, http.StatusInternalServerError)
			log.Err(err).Msg("cannot get products")
			return
		}
		products = append(products, product)
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
	res, err := postgres.DB.Exec("UPDATE orders SET status = $1 WHERE id=$2", st.Status, id)
	if err != nil {
		http.Error(w, `{"statusCode": 500, "details": "CANNOT_UPDATE_ORDER"}`, http.StatusInternalServerError)
		log.Err(err).Msg("cannot update order status")
		return
	}

	n, _ := res.RowsAffected()
	if  n == 0 {
		http.Error(w, `{"statusCode": 404, "details": "ORDER_NOT_FOUND"}`, http.StatusNotFound)
		log.Err(errors.New("order not found")).Msg("cannot update order status")
		return
	}

	resp := fmt.Sprintf(`{"order_id": "%s", "status": "%s", "details": "updated"}`, id, st.Status)
	_, _ = w.Write([]byte(resp))
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
	log.Info().Msgf("App starting... Listen: %s\n", address)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Err(err).Msg("problem with starting app")
		return
	}

}
