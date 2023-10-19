package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	LogsTopic     = "log"
	MessagesTopic = "message"

	EnvVarRabbitMqAddr = "RABBITMQ_ADDR"
)

func main() {
	// Lookup service 2 address or default to expected value.
	svc2Address, ok := os.LookupEnv("SERVICE2")
	if !ok {
		svc2Address = "service2"
	}
	fullSvc2Address := svc2Address + ":8000"

	rabbitmqAddr := os.Getenv(EnvVarRabbitMqAddr)
	conn, err := amqp.Dial(rabbitmqAddr)
	for err != nil {
		logrus.Warn("failed to connect; retrying in 2 seconds")
		<-time.After(2 * time.Second)

		conn, err = amqp.Dial(rabbitmqAddr)
	}
	defer conn.Close()

	logsPub, err := NewPublisher(conn, LogsTopic)
	if err != nil {
		logrus.Fatal("failed to create publisher:", err)
	}
	defer logsPub.Close()

	msgsPub, err := NewPublisher(conn, MessagesTopic)
	if err != nil {
		logrus.Fatal("failed to create publisher:", err)
	}
	defer msgsPub.Close()

	// Send 20 texts to service 2.
	for i := 1; i <= 20; i++ {
		addresses, err := resolveAddresses(fullSvc2Address)
		if err != nil {
			logsPub.Publish(err.Error())
		} else {
			timestamp := timestampNow()
			line := fmt.Sprintf("SND %d %v %s", i, timestamp, addresses.tcpAddr)

			// Send message to 'message' topic.
			if err := msgsPub.Publish(line); err != nil {
				// Report error.
				logsPub.Publish(err.Error())
			}

			// Send message via HTTP to service 2.
			resp, err := http.Post(addresses.httpAddr, "text/plain", strings.NewReader(line))
			if err != nil {
				// Report error.
				logsPub.Publish(err.Error())
			} else {
				// Send response code and timestamp to 'log' topic.
				logsPub.Publish(fmt.Sprintf("%d %s", resp.StatusCode, timestamp))
			}
		}

		// Wait 2 seconds between requests.
		<-time.After(2 * time.Second)
	}

	// Send stop signal.
	logsPub.Publish("SND STOP")

	// Wait until program is terminated.
	var wait chan struct{}
	<-wait
}

func timestampNow() string {
	return time.Now().UTC().Round(time.Millisecond).Format(time.RFC3339Nano)
}

// Container for TCP address and its corresponding HTTP address.
type Addresses struct {
	tcpAddr  *net.TCPAddr
	httpAddr string
}

// Resolves domain name to TCP and HTTP addresses.
func resolveAddresses(serviceAddress string) (*Addresses, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", serviceAddress)
	if err != nil {
		return nil, err
	}

	httpAddr := fmt.Sprintf("http://%v/", tcpAddr)

	return &Addresses{
		tcpAddr:  tcpAddr,
		httpAddr: httpAddr,
	}, nil
}
