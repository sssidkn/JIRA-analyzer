FROM golang:1.23.3 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /analytics-service .
COPY --from=builder /app/config/config.yaml ./config/config.yaml

EXPOSE 8084
CMD ["./analytics-service"]