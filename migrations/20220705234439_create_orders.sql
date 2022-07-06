-- +goose Up
-- +goose StatementBegin
CREATE TYPE status AS ENUM ('new', 'confirmed', 'done', 'canceled');
CREATE TABLE orders
(
    id    bigserial primary key,
    address varchar(300),
    name varchar(300),
    phone varchar(300),
    status status
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders
-- +goose StatementEnd
