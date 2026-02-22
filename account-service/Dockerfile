# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /account-service ./cmd/account

# Run stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /account-service .

EXPOSE 8080

# Pass env at runtime: DATABASE_URL (required), PORT, SECRET_KEY, ISSUER
ENTRYPOINT ["./account-service"]
