FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN go build -o /btc-rate-api

EXPOSE 8080

CMD [ "/btc-rate-api" ]