# Stage 1: Build phase
FROM golang:1.24.2-alpine AS builder
# Cài đặt các dependencies cần thiết
RUN apk update && apk add --no-cache git
WORKDIR /app
# Chỉ copy các file cần thiết để build
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
# Build ứng dụng
RUN go build -ldflags="-s -w" -o myapp .
# Stage 2: Final image with just the binary
FROM alpine:latest
WORKDIR /root/
# Copy binary vào image cuối
COPY --from=builder /app/myapp .
# Chạy ứng dụng
CMD ["./myapp"]

