# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o microservice .

# Run stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/microservice .
EXPOSE 8080
CMD ["./microservice"]