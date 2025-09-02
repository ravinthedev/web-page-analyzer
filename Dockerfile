FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /root/

COPY --from=builder /app/api .

COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./api"]
