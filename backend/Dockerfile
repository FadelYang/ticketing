# Stage 1: Build Go API server and migration CLI
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate ./cmd/migrate \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/seed ./cmd/seed

# Stage 2: Minimal runtime image
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/seed /app/seed

EXPOSE 8080

CMD ["/app/server"]
