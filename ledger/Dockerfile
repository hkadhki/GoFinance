FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./cmd

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/app .

RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser
USER appuser

CMD ["./app"]