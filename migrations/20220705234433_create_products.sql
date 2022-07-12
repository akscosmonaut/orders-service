-- +goose Up
-- +goose StatementBegin
CREATE TABLE products
(
    id    bigserial primary key,
    name varchar(256) unique,
    price int
);
/*
INSERT INTO products (id, name, price)
VALUES (1, 'Хлеб Коломенский', 58),
       (2, 'Молоко Домик в деревне', 80),
       (3, 'Сыр Гауда', 200),
       (4, 'Маслины Iberica', 150),
       (5, 'Жевательная резинка Orbit', 30),
       (6, 'Йогурт ЧУДО', 70),
       (7, 'Шоколад Alpen Gold', 95);*/
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE products;
-- +goose StatementEnd
