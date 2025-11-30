FROM golang:1.23.3 AS builder

WORKDIR /app
COPY . .

RUN go mod download

RUN go test -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /connector-service ./cmd/service/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /connector-service .

EXPOSE 9090 8081
CMD ["./connector-service"]