#!/bin/sh

#echo "START"
#echo $ORDERS_SERVICE_DB
#goose -dir ../migrations postgres "$ORDERS_SERVICE_DB" up
##docker-compose up --build


echo "START"
source .env
echo "STARTING DATABASE AND BACKEND SERVICE"
docker-compose up --build -d
echo "WAITING DATABASE"
sleep 5
echo "RUN MIGRATION FOR DATABASE $GOOSE_ORDERS_SERVICE_DB"
goose -dir ../migrations postgres "$GOOSE_ORDERS_SERVICE_DB" up
docker-compose logs -f orders-service