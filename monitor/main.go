package main

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	LogsTopic = "log"

	EnvVarRabbitMqAddr = "RABBITMQ_ADDR"
)

// Thread safe container for all received log lines.
type LogStorage struct {
	lck  sync.RWMutex
	logs string
}

// Push new log line to buffer.
func (s *LogStorage) PushLine(line string) {
	s.lck.Lock()
	defer s.lck.Unlock()

	s.logs += line + "\n"
}

// Get all received logs.
func (s *LogStorage) Get() string {
	s.lck.RLock()
	defer s.lck.RUnlock()

	return s.logs
}

func main() {
	rabbitmqAddr := os.Getenv(EnvVarRabbitMqAddr)

	// Retry until RabbitMQ connection is established.
	conn, err := amqp.Dial(rabbitmqAddr)
	for err != nil {
		logrus.Warn("failed to connect; retrying in 2 seconds")
		<-time.After(2 * time.Second)

		conn, err = amqp.Dial(rabbitmqAddr)
	}
	defer conn.Close()

	// Create new topic subscriber.
	subscriber, err := NewSubscriber(conn)
	if err != nil {
		logrus.Fatal("failed to create a subscriber: ", err)
	}
	defer subscriber.Close()

	// Get channel to receive topic messages.
	logMsgs, err := subscriber.Channel()
	if err != nil {
		logrus.Fatal("failed to get subscriber channel: ", err)
	}

	// Container for storing logs.
	storage := &LogStorage{}

	// Start log listener.
	go func() {
		for logMsg := range logMsgs {
			storage.PushLine(string(logMsg.Body))
		}
	}()

	logrus.Info("listener running")

	// Start server for GETting logs.
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, storage.Get())
	})

	go router.Run(":8087")

	logrus.Info("server running")

	// Wait until program is terminated.
	var wait chan struct{}
	<-wait
}
