package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Lookup service 2 address or default to expected value.
	svc2Address, ok := os.LookupEnv("SERVICE2")
	if !ok {
		svc2Address = "service2"
	}
	fullSvc2Address := svc2Address + ":8000"

	// Create service 1 specific log file.
	logFile, err := createLogFile()
	if err != nil {
		log.Fatal("failed to create log file: ", err)
	}
	defer logFile.Close()

	// Send 20 texts to service 2.
	for i := 1; i <= 20; i++ {
		addresses, err := resolveAddresses(fullSvc2Address)
		if err != nil {
			logFile.WriteString(fmt.Sprintln(err.Error()))
		} else {
			timestamp := time.Now().UTC().Round(time.Millisecond).Format(time.RFC3339Nano)
			line := fmt.Sprintf("%d %v %s", i, timestamp, addresses.tcpAddr)

			logAndPost(line, addresses.httpAddr, logFile)
		}

		// Wait 2 seconds between requests.
		<-time.After(2 * time.Second)
	}

	addresses, err := resolveAddresses(fullSvc2Address)
	if err != nil {
		logFile.WriteString(fmt.Sprintln(err.Error()))
		return
	}

	// Stop communication by sending signal.
	logAndPost("STOP", addresses.httpAddr, logFile)
}

func createLogFile() (*os.File, error) {
	err := os.MkdirAll("./logs", os.ModePerm)
	if err != nil {
		return nil, err
	}
	logFile, err := os.Create("logs/service1.log")
	if err != nil {
		return nil, err
	}

	return logFile, nil
}

// Helper to log and post the next line.
func logAndPost(line, httpAddr string, file *os.File) {
	file.WriteString(line + "\n")

	err := postString(httpAddr, line)
	if err != nil {
		file.WriteString(fmt.Sprintln(err.Error()))
	}
}

// Do a POST request at given HTTP address with given body string.
func postString(httpAddr, str string) error {
	reader := strings.NewReader(str)

	resp, err := http.Post(httpAddr, "text/plain", reader)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

// Container for TCP address corresponding HTTP address.
type Addresses struct {
	tcpAddr  *net.TCPAddr
	httpAddr string
}

// Resolve domain name to TCP and HTTP addresses.
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