package main

import (
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	LogsTopic = "logs"

	EnvVarRabbitMqAddr = "RABBITMQ_ADDR"
)

type LogStorage struct {
	lck  sync.RWMutex
	logs string
}

func (s *LogStorage) PushLine(line string) {
	s.lck.Lock()
	defer s.lck.Unlock()

	s.logs += line + "\n"
}

func (s *LogStorage) Get() string {
	s.lck.RLock()
	defer s.lck.RUnlock()

	return s.logs
}

func main() {
	rabbitmqAddr := os.Getenv(EnvVarRabbitMqAddr)

	subscriber, err := NewSubscriber(rabbitmqAddr)
	if err != nil {
		logrus.Fatal("failed to create a subscriber: ", err)
	}

	msgs, err := subscriber.Channel()
	if err != nil {
		logrus.Fatal("failed to get subscriber channel: ", err)
	}

	// Container for storing logs.
	storage := &LogStorage{}

	// Start log listener.
	go func() {
		for msg := range msgs {
			storage.PushLine(string(msg.Body))
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
