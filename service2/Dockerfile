# Build from source
FROM rust:1.72.0 AS builder
WORKDIR /build
COPY Cargo.lock Cargo.toml ./
COPY src/* ./src/
RUN cargo install --path .

# Install
FROM rust:1.72.0-slim AS release
WORKDIR /app
COPY --from=builder /build/target/release/service2 ./svc2

EXPOSE 8000

CMD ["/app/svc2"]
