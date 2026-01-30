FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the dashboard command. Note: main.go and token_handlers.go are in the same package.
RUN go build -o dashboard ./cmd/dashboard

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/dashboard .
EXPOSE 8080
CMD ["./dashboard"]