-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders_products
(
    id    bigserial primary key unique,
    order_id int,
    product_id int
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders_products;
-- +goose StatementEnd
