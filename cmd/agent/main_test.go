package main

import (
	"bytes"
	"flag"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestEnvVariables checks if environment variables work correctly.
func TestEnvVariables(t *testing.T) {
	t.Setenv("ADDRESS", "envhost:9090")
	t.Setenv("REPORT_INTERVAL", "15")
	t.Setenv("POLL_INTERVAL", "5")

	flagAddr = "flaghost:8081"
	flagReportInt = 10
	flagPollInt = 2

	address := getEnvOrFlag("ADDRESS", flagAddr, "localhost:8080")
	reportInterval := time.Duration(getIntEnvOrFlag("REPORT_INTERVAL", flagReportInt, 10)) * time.Second
	pollInterval := time.Duration(getIntEnvOrFlag("POLL_INTERVAL", flagPollInt, 2)) * time.Second

	if address != "envhost:9090" {
		t.Errorf("Expected address to be 'envhost:9090' from environment variable, got %s", address)
	}
	if reportInterval != 15*time.Second {
		t.Errorf("Expected reportInterval to be 15 seconds from environment variable, got %v", reportInterval)
	}
	if pollInterval != 5*time.Second {
		t.Errorf("Expected pollInterval to be 5 seconds from environment variable, got %v", pollInterval)
	}
}

// TestFlagDefaults checks that default flag values are set correctly.
func TestFlagDefaults(t *testing.T) {
	testFlags := flag.NewFlagSet("test_flags", flag.ExitOnError)
	var testFlagAddr string
	var testFlagReportInt, testFlagPollInt int

	// Define flags in the new FlagSet
	testFlags.StringVar(&testFlagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	testFlags.IntVar(&testFlagReportInt, "r", 10, "Report interval in seconds")
	testFlags.IntVar(&testFlagPollInt, "p", 2, "Poll interval in seconds")

	// Parse the flags (no custom values are provided, so defaults are used)
	if err := testFlags.Parse([]string{}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if testFlagAddr != "localhost:8080" {
		t.Errorf("Expected default address to be 'localhost:8080', got %s", testFlagAddr)
	}
	if testFlagReportInt != 10 {
		t.Errorf("Expected default report interval to be 10, got %d", testFlagReportInt)
	}
	if testFlagPollInt != 2 {
		t.Errorf("Expected default poll interval to be 2, got %d", testFlagPollInt)
	}
}

// TestInvalidFlags checks that the program exits correctly when unknown flags are provided.
func TestInvalidFlags(t *testing.T) {
	var buf bytes.Buffer

	// Create a new FlagSet for testing unknown flags
	testFlags := flag.NewFlagSet("test_invalid_flags", flag.ContinueOnError)
	testFlags.SetOutput(&buf)

	// Define flags in the new FlagSet
	testFlags.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	testFlags.IntVar(&flagReportInt, "r", 10, "Report interval in seconds")
	testFlags.IntVar(&flagPollInt, "p", 2, "Poll interval in seconds")

	// Parse an unknown flag
	err := testFlags.Parse([]string{"-unknown"})

	// Check if error contains "unknown flag" warning
	if err == nil || !bytes.Contains(buf.Bytes(), []byte("flag provided but not defined")) {
		t.Error("Expected error message for unknown flag")
	}
}

// TestNewMetrics checks the initialization of the Metrics structure.
func TestNewMetrics(t *testing.T) {
	metrics := newMetrics()

	if len(metrics.Gauges) == 0 {
		t.Error("Expected non-empty Gauges map")
	}
	if len(metrics.Counters) == 0 {
		t.Error("Expected non-empty Counters map")
	}

	// Check that specific keys are present in metrics
	if _, ok := metrics.Gauges["Alloc"]; !ok {
		t.Error("Expected Alloc metric in Gauges")
	}
	if _, ok := metrics.Counters["PollCount"]; !ok {
		t.Error("Expected PollCount metric in Counters")
	}
}

// TestUpdateMetrics checks that metric values are updated correctly.
func TestUpdateMetrics(t *testing.T) {
	metrics := newMetrics()
	metrics.updateMetrics()

	// Check that some metrics have been updated to non-zero values
	if metrics.Gauges["Alloc"] == 0 {
		t.Error("Expected Alloc to be updated to a non-zero value")
	}
	if metrics.Counters["PollCount"] != 1 {
		t.Errorf("Expected PollCount to increment, got %d", metrics.Counters["PollCount"])
	}
	if metrics.Gauges["RandomValue"] < 0 || metrics.Gauges["RandomValue"] > 100 {
		t.Errorf("Expected RandomValue in range 0-100, got %v", metrics.Gauges["RandomValue"])
	}
}

// TestSendMetric checks that the metric sending function works correctly.
func TestSendMetric(t *testing.T) {
	metrics := newMetrics()

	// Create a test server to check metric sending
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "gauge/Alloc/") {
			t.Errorf("Expected URL path to contain 'gauge/Alloc/', got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected Content-Type to be text/plain, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	destAddress := serverURL.Host
	metrics.sendMetric(destAddress, "gauge", "Alloc", strconv.FormatFloat(rand.Float64()*100, 'f', -1, 64))
}

// TestReportMetrics checks the sending of all metrics.
func TestReportMetrics(t *testing.T) {
	metrics := newMetrics()
	metrics.updateMetrics()

	// Counter to check that all metrics are sent
	counter := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	destAddress := serverURL.Host

	metrics.reportMetrics(destAddress)

	// Check that the expected number of metrics were sent
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reportDone <- true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	destAddress := serverURL.Host

	// Override flag values for polling and reporting intervals.
	flagPollInt = 1   // poll interval in seconds.
	flagReportInt = 3 // report interval in seconds.

	go func() {
		metrics.updateMetrics()
		pollDone <- true
	}()

	go func() {
		metrics.reportMetrics(destAddress)
	}()

	select {
	case <-pollDone:
	case <-time.After(2 * time.Duration(flagPollInt) * time.Second):
		t.Errorf("Poll metrics function timed out")
	}

	select {
	case <-reportDone:
	case <-time.After(2 * time.Duration(flagReportInt) * time.Second):
		t.Errorf("Report metrics function timed out")
	}
}
