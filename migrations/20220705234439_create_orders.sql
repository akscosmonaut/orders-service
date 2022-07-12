-- +goose Up
-- +goose StatementBegin
CREATE TYPE status AS ENUM ('new', 'confirmed', 'done', 'canceled');
CREATE TABLE orders
(
    id bigserial primary key,
    address varchar(256),
    name varchar(256),
    phone varchar(256),
    status status,
    products bigint[],
    total int

);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE status;
DROP TABLE orders;
-- +goose StatementEnd
