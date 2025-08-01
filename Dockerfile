# Stage 1: Build the Go app
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o web-image main.go

# Stage 2: Alpine with certs
FROM alpine:latest

WORKDIR /app

# ✅ Fix: Add root certs for HTTPS
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/web-image .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./web-image"]
