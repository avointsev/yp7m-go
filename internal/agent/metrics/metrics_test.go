package metrics

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestNewMetrics checks the initialization of the Metrics structure.
func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

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
	metrics := NewMetrics()
	metrics.UpdateMetrics()

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
	metrics := NewMetrics()

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
	metrics.SendMetric(destAddress, "gauge", "Alloc", strconv.FormatFloat(rand.Float64()*100, 'f', -1, 64))
}

// TestReportMetrics checks the sending of all metrics.
func TestReportMetrics(t *testing.T) {
	metrics := NewMetrics()
	metrics.UpdateMetrics()

	// Counter to check that all metrics are sent
	counter := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	destAddress := serverURL.Host

	metrics.ReportMetrics(destAddress)

	// Check that the expected number of metrics were sent
	expectedCount := len(metrics.Gauges) + len(metrics.Counters)
	if counter != expectedCount {
		t.Errorf("Expected %d metrics to be reported, got %d", expectedCount, counter)
	}
}

// TestMainLoop Test the main function using timers.
func TestMainLoop(t *testing.T) {
	metrics := NewMetrics()

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
	flagPollInt := 1   // poll interval in seconds.
	flagReportInt := 3 // report interval in seconds.

	go func() {
		metrics.UpdateMetrics()
		pollDone <- true
	}()

	go func() {
		metrics.ReportMetrics(destAddress)
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
