package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PLAIN_CONTENT_TYPE = "text/plain"
)

func main() {
	// Create service 1 specific log file.
	newpath := filepath.Join(".", "logs")
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Fatal("failed to create directory: ", err)
	}
	logFile, err := os.Create("logs/service1.log")
	if err != nil {
		log.Fatal("failed to create file: ", err)
	}
	defer logFile.Close()

	serverAddress := "127.0.0.1:8000"
	httpAddress := fmt.Sprintf("http://%s", serverAddress)

	// Send 20 texts to service 2.
	for i := 1; i <= 20; i++ {
		timestamp := time.Now().UTC().Round(time.Millisecond).Format(time.RFC3339Nano)
		line := fmt.Sprintf("%d %v %s", i, timestamp, serverAddress)

		logFile.WriteString(fmt.Sprintf("%s\n", line))
		reader := strings.NewReader(line)

		_, err = http.Post(httpAddress, PLAIN_CONTENT_TYPE, reader)
		if err != nil {
			logFile.WriteString(fmt.Sprintln(err.Error()))
		}

		// Wait 2 seconds between requests.
		<-time.After(2 * time.Second)
	}

	logFile.WriteString("STOP\n")

	// Stop communication by sending signal.
	reader := strings.NewReader("STOP")
	http.Post(httpAddress, PLAIN_CONTENT_TYPE, reader)
}
