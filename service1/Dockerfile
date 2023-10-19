# Build from source
FROM golang:1.21-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./svc1

# Install
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /build/svc1 /usr/bin/svc1

CMD ["/bin/sh", "-c", "svc1"]
