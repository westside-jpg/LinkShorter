# Этап сборки
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o linkshorter main.go

# Финальный этап, кладём в лёгкий образ только готовый бинарник
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/linkshorter .

EXPOSE 8080

CMD ["./linkshorter"]
