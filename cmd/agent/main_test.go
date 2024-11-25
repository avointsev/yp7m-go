package main

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/avointsev/yp7m-go/internal/agent/metrics"
	"github.com/avointsev/yp7m-go/internal/flags"
)

func TestMainFunction(t *testing.T) {
	config := flags.AgentConfig{
		ReportInterval: 2 * time.Second,
		PollInterval:   1 * time.Second,
		Address:        "http://localhost:8080",
	}

	mockMetrics := metrics.NewMetrics()

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	tickerPoll := time.NewTicker(config.PollInterval)
	tickerReport := time.NewTicker(config.ReportInterval)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(5 * time.Second)
		done <- true
	}()

	// Run the main loop in a separate goroutine
	go func() {
		for {
			select {
			case <-tickerPoll.C:
				mockMetrics.UpdateMetrics()
			case <-tickerReport.C:
				mockMetrics.ReportMetrics(config.Address)
			}
		}
	}()

	// Wait for the test to complete
	<-done

	logOutput := logBuffer.String()
	if logOutput != "" {
		t.Errorf("unexpected log output: %v", logOutput)
	}
}
