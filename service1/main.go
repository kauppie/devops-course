package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PLAIN_CONTENT_TYPE = "text/plain"
	SERVICE2_ADDRESS   = "service2:8000"
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

	// Send 20 texts to service 2.
	for i := 1; i <= 20; i++ {
		tcpAddr, err := net.ResolveTCPAddr("tcp", SERVICE2_ADDRESS)
		if err != nil {
			logFile.WriteString(fmt.Sprintln(err.Error()))

			// Wait 2 seconds between requests.
			<-time.After(2 * time.Second)
			continue
		}
		httpAddress := fmt.Sprintf("http://%v/", tcpAddr)

		timestamp := time.Now().UTC().Round(time.Millisecond).Format(time.RFC3339Nano)
		line := fmt.Sprintf("%d %v %s", i, timestamp, tcpAddr)

		logFile.WriteString(fmt.Sprintf("%s\n", line))
		reader := strings.NewReader(line)

		resp, err := http.Post(httpAddress, PLAIN_CONTENT_TYPE, reader)
		if err != nil {
			logFile.WriteString(fmt.Sprintln(err.Error()))
		} else {
			resp.Body.Close()
		}

		// Wait 2 seconds between requests.
		<-time.After(2 * time.Second)
	}

	logFile.WriteString("STOP\n")

	tcpAddr, err := net.ResolveTCPAddr("tcp", SERVICE2_ADDRESS)
	if err != nil {
		logFile.WriteString(fmt.Sprintln(err.Error()))
		return
	}
	httpAddress := fmt.Sprintf("http://%v/", tcpAddr)

	// Stop communication by sending signal.
	reader := strings.NewReader("STOP")
	http.Post(httpAddress, PLAIN_CONTENT_TYPE, reader)
}
