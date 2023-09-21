FROM golang:1.21.1 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o microgopster

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=builder /app/microgopster /microgopster

ENTRYPOINT [ "/microgopster" ]