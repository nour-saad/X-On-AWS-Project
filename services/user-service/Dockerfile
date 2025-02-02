FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service

FROM alpine:latest
WORKDIR /app
COPY --from=builder /user-service /app/user-service
EXPOSE 8080
CMD ["/app/user-service"]