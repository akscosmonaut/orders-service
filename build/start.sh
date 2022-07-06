#!/bin/sh

echo "START"
source .env
cd ..
echo "MIGRATIONS RUN"
echo $ORDERS_SERVICE_DB
goose -dir migrations postgres "$ORDERS_SERVICE_DB" up
