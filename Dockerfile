FROM golang:1.21.1 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go build .

ENTRYPOINT [ "/microgopster" ]