# Giai đoạn 1: Build ứng dụng
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY env/ env/
COPY internal/templates/ internal/templates/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service ./cmd/http-server/main.go

# Giai đoạn 2: Runtime
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/auth-service .
COPY --from=builder /app/env ./env
COPY --from=builder /app/internal/templates ./internal/templates

CMD ["./auth-service"]
