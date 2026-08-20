FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build statically linked binary and strip debug info (-s -w) for smaller size
RUN CGO_ENABLED=0 GOOS=linux go build -p 1 -ldflags="-s -w" -o main ./cmd/server/main.go

FROM alpine:latest

# tzdata cần cho DSN TimeZone=Asia/Ho_Chi_Minh (Postgres driver phân giải timezone lúc kết nối).
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
