#!/bin/sh

echo "START"
source .env
echo "STARTING DATABASE AND BACKEND SERVICE"
docker-compose up --build -d
echo "RUN MIGRATION FOR DATABASE $GOOSE_ORDERS_SERVICE_DB"
goose -dir ../migrations postgres "$GOOSE_ORDERS_SERVICE_DB" up
docker-compose logs -f orders-service
echo "TEST RUN"
cd ../tests
go test . -count=1 "$@"