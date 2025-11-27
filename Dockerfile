# # Stage 1: Build phase
# FROM golang:alpine AS builder
# # Cài đặt các dependencies cần thiết
# RUN apk update && apk add --no-cache git
# WORKDIR /app
# # Chỉ copy các file cần thiết để build
# COPY go.mod go.sum ./
# RUN go mod tidy
# COPY . .
# # Build ứng dụng
# RUN go build -ldflags="-s -w" -o myapp .
# # Stage 2: Final image with just the binary
# FROM alpine:latest
# WORKDIR /root/
# # Copy binary vào image cuối
# COPY --from=builder /app/myapp .
# # Chạy ứng dụng
# CMD ["./myapp"]
# =========================
# Stage 1: Build phase
# =========================
# =========================
# Stage 1: Build phase
# =========================
FROM golang:1.22-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

# Bật modules + toolchain auto
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOTOOLCHAIN=auto

WORKDIR /app

# Copy go.mod + go.sum trước để cache dependency
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ source
COPY . .

# Build ứng dụng
RUN go build -ldflags="-s -w" -o myapp .

# =========================
# Stage 2: Final image
# =========================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/myapp .

EXPOSE 8080

CMD ["./myapp"]



