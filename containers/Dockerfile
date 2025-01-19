# Stage 1: Build
FROM golang:1.20 as builder
WORKDIR /app
COPY main.go .
RUN go mod init twitter-service && go mod tidy
RUN go build -o twitter-service .

# Stage 2: Runtime
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/twitter-service .
EXPOSE 8080
ENV PORT=8080
CMD ["./twitter-service"]
