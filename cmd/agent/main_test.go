package main

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

// TestUpdateMetrics Test updating metrics using data from runtime and random values.
func TestUpdateMetrics(t *testing.T) {
	metrics := newMetrics()
	metrics.updateMetrics()

	// Check that gauge metrics are updated (values are non-zero).
	if metrics.Gauges["Alloc"] == 0 {
		t.Errorf("Expected Alloc to be updated, got 0")
	}
	if metrics.Gauges["RandomValue"] < 0 || metrics.Gauges["RandomValue"] > 100 {
		t.Errorf("Expected RandomValue to be in range 0-100, got %v", metrics.Gauges["RandomValue"])
	}

	// Check that PollCount counter increments.
	initialPollCount := metrics.Counters["PollCount"]
	metrics.updateMetrics()
	if metrics.Counters["PollCount"] != initialPollCount+1 {
		t.Errorf("Expected PollCount to increment by 1, got %v", metrics.Counters["PollCount"])
	}
}

// TestSendMetric Test sending a metric to the server using a mock server.
func TestSendMetric(t *testing.T) {
	metrics := newMetrics()

	// Create a test server that checks request correctness.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check the header.
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type to be text/plain, got %s", r.Header.Get("Content-Type"))
		}

		// Check the URL format.
		expectedPath := "/update/gauge/Alloc/"
		if r.URL.Path[:len(expectedPath)] != expectedPath {
			t.Errorf("Expected URL path to start with %s, got %s", expectedPath, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Error parsing server URL: %v", err)
	}
	host = serverURL.Hostname()
	port = serverURL.Port()

	metrics.sendMetric("gauge", "Alloc", strconv.FormatFloat(rand.Float64()*100, 'f', -1, 64))
}

// TestReportMetrics Test sending all metrics to the server.
func TestReportMetrics(t *testing.T) {
	metrics := newMetrics()
	metrics.updateMetrics()

	counter := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Error parsing server URL: %v", err)
	}
	host = serverURL.Hostname()
	port = serverURL.Port()

	metrics.reportMetrics()

	// Check that the expected number of metrics were sent.
	expectedCount := len(metrics.Gauges) + len(metrics.Counters)
	if counter != expectedCount {
		t.Errorf("Expected %d metrics to be reported, got %d", expectedCount, counter)
	}
}

// TestMainLoop Test the main function using timers.
func TestMainLoop(t *testing.T) {
	metrics := newMetrics()

	pollDone := make(chan bool)
	reportDone := make(chan bool)

	// Start a server to check report sending.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportDone <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Parse the server URL and set host and port.
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Error parsing server URL: %v", err)
	}
	host = serverURL.Hostname()
	port = serverURL.Port()

	// Override intervals.
	pollInterval = 1 * time.Second
	reportInterval = 3 * time.Second

	go func() {
		metrics.updateMetrics()
		pollDone <- true
	}()

	go func() {
		metrics.reportMetrics()
	}()

	select {
	case <-pollDone:
	case <-time.After(2 * pollInterval):
		t.Errorf("Poll metrics function timed out")
	}

	select {
	case <-reportDone:
	case <-time.After(2 * reportInterval):
		t.Errorf("Report metrics function timed out")
	}
}
