FROM golang:1.25.4

WORKDIR /app
COPY tests/integration/integration_test.go .
COPY tests/go.mod ./
RUN go mod download
RUN go test -v integration_test.go
