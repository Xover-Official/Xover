FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the dashboard command. Note: main.go and token_handlers.go are in the same package.
RUN go build -o dashboard ./cmd/dashboard

FROM alpine:latest
WORKDIR /root/
# Install curl for health checks
RUN apk --no-cache add curl
COPY --from=builder /app/dashboard .
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

CMD ["./dashboard"]