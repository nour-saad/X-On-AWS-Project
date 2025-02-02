FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /tweet-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /tweet-service /app/tweet-service
EXPOSE 8081
CMD ["/app/tweet-service"]