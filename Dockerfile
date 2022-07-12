FROM library/golang:1.14-alpine

RUN apk -U add bash make

ENV APP_DIR $GOPATH/src/orders-service
WORKDIR $APP_DIR
COPY . .

ADD https://github.com/pressly/goose/releases/download/v2.6.0/goose-linux64 /bin/goose
RUN chmod +x /bin/goose
ARG ORDERS_SERVICE_DB
RUN /bin/goose -dir migrations postgres ${ORDERS_SERVICE_DB} up

RUN GOOS=linux go build -mod vendor -ldflags "-w -s " -o orders-service .
CMD (./orders-service)

EXPOSE 9000