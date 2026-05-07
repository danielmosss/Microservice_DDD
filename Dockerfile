# Build stage
FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . .
RUN swag init
RUN go build -o DDD_Microservice .

# Run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app .
EXPOSE 8080
CMD ["./DDD_Microservice"]