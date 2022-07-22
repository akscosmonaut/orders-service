-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders_products
(
    id    bigserial primary key unique,
    order_id int,
    product_id int
);

INSERT INTO orders_products (order_id, product_id)
VALUES (1, 2),
       (1, 6),
       (1, 8),
       (2, 1),
       (2, 10),
       (2, 15),
       (2, 16),
       (2, 20),
       (3, 9),
       (3, 11);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders_products;
-- +goose StatementEnd
