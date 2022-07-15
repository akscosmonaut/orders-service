FROM library/golang:1.14-alpine

RUN apk -U add bash make

ENV APP_DIR $GOPATH/src/orders-service
WORKDIR $APP_DIR
COPY . .

RUN go mod vendor
RUN GOOS=linux go build -mod vendor -ldflags "-w -s " -o orders-service .
CMD (./orders-service)

EXPOSE 9000