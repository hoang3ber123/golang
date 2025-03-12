# Giai đoạn 1: Build ứng dụng
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY env/ env/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./cmd/http-server/main.go

# Giai đoạn 2: Runtime
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/product-service .
COPY --from=builder /app/env ./env

CMD ["./product-service"]
