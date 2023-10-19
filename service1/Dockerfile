# Build from source
FROM golang:1.21.3-bookworm AS builder
WORKDIR /build
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./svc1

# Install
FROM alpine:3.18 AS release
WORKDIR /app
COPY --from=builder /build/svc1 /usr/bin/svc1

CMD ["/bin/sh", "-c", "svc1"]
