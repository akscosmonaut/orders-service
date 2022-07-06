-- +goose Up
-- +goose StatementBegin
CREATE TABLE products
(
    id    bigserial primary key,
    name varchar(300) unique,
    price int
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE products
-- +goose StatementEnd
