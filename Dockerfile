FROM golang:1.21 AS builder

WORKDIR /build

COPY helloserver/go.mod ./
RUN go mod download

# Build helloserver
COPY helloserver/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ../helloserver

FROM ubuntu:jammy

# Update registry
RUN apt-get update

# Install SSH server
RUN apt-get install -y openssh-server
RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config

# Create user
RUN useradd -m -s /bin/bash -G sudo -p $(openssl passwd -1 eee) ssluser

# Install tools
RUN apt-get install -y python3 sudo net-tools

# Install HTTP helloserver
COPY --from=builder /helloserver /usr/bin/helloserver

# Install SSH key
COPY key.pub /home/ssluser/.ssh/authorized_keys

EXPOSE 8080

ENTRYPOINT service ssh start && helloserver
