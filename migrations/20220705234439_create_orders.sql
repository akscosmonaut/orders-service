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
    total int

);

INSERT INTO orders (address, name, phone, status, total)
VALUES ('Бассейная, 57', 'Роман', '+78339992211', 'new', 712),
       ('Дубравы, 67', 'Сергей', '+78334442211', 'new', 910),
       ('Мира, 203', 'Татьяна', '+78339388311', 'confirmed', 645);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TYPE status;
DROP TABLE orders;
-- +goose StatementEnd
