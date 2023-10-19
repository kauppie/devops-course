# Build from source
FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ../monitor

# Install
FROM alpine:3.18

WORKDIR /app

COPY --from=builder /monitor /usr/bin/monitor

EXPOSE 8087

CMD ["monitor"]
