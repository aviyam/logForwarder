# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o logforwarder main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/logforwarder .

EXPOSE 5044 24224

ENV LUMBERJACK_ADDR=":5044"
ENV FLUENT_ADDR=":24224"

CMD ["./logforwarder"]
