# Build from source
FROM rust:1.72 AS builder

# Install tools required for musl platform
RUN apt-get update \
  && apt-get install --no-install-recommends -y musl-tools
RUN rustup target add x86_64-unknown-linux-musl

WORKDIR /build

COPY . ./

RUN cargo install --target x86_64-unknown-linux-musl --path .

# Install
FROM alpine:3.18 AS release
WORKDIR /app
COPY --from=builder /build/target/x86_64-unknown-linux-musl/release/service2 /usr/bin/svc2

EXPOSE 8000

CMD ["/bin/sh", "-c", "svc2"]
