package main

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestEnvVariables checks if environment variable works correctly.
func TestEnvVariables(t *testing.T) {
	t.Setenv("ADDRESS", "envhost:9090")
	flagAddr = "flaghost:8081"

	address := getEnvOrFlag("ADDRESS", flagAddr, "localhost:8080")

	if address != "envhost:9090" {
		t.Errorf("Expected address to be 'envhost:9090' from environment variable, got %s", address)
	}
}

// setupRouter creates and returns a new chi router for testing.
func setupRouter(storage Storage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", rootHandler(storage))
	r.Get("/value/{type}/{name}", getMetricHandler(storage))
	r.Post("/update/{type}/{name}/{value}", updateMetricHandler(storage))
	return r
}

// TestFlagDefaults checks that default flag values are set correctly.
func TestFlagDefaults(t *testing.T) {
	testFlags := flag.NewFlagSet("test_flags", flag.ExitOnError)
	var testFlagAddr string

	testFlags.StringVar(&testFlagAddr, "a", "localhost:8080", "HTTP server endpoint address")

	if err := testFlags.Parse([]string{}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if testFlagAddr != "localhost:8080" {
		t.Errorf("Expected default address to be 'localhost:8080', got %s", testFlagAddr)
	}
}

// TestInvalidFlags checks that the program exits correctly when unknown flags are provided.
func TestInvalidFlags(t *testing.T) {
	var buf bytes.Buffer

	testFlags := flag.NewFlagSet("test_invalid_flags", flag.ContinueOnError)
	testFlags.SetOutput(&buf)

	testFlags.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	err := testFlags.Parse([]string{"-unknown"})

	if err == nil || !bytes.Contains(buf.Bytes(), []byte("flag provided but not defined")) {
		t.Error("Expected error message for unknown flag")
	}
}

// TestRootHandler tests the root handler.
func TestRootHandler(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateGauge("testGauge", 123.45)
	storage.UpdateCounter("testCounter", 100)

	r := setupRouter(storage)
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	expected := "<html>"
	if !strings.Contains(string(body), expected) {
		t.Errorf("expected body to contain %q; got %q", expected, body)
	}
}

// TestUpdateMetricHandler tests updating a metric.
func TestUpdateMetricHandler(t *testing.T) {
	storage := NewMemStorage()
	r := setupRouter(storage)
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/123.45", http.NoBody)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	expected := "Metric testGauge updated successfully"
	if !strings.Contains(string(body), expected) {
		t.Errorf("expected body to contain %q; got %q", expected, body)
	}

	// Verify that the metric value has been updated.
	if storage.GetAllMetrics()["testGauge"] != 123.45 {
		t.Errorf("expected testGauge to be 123.45; got %v", storage.GetAllMetrics()["testGauge"])
	}
}

// TestGetMetricHandler tests retrieving a metric value.
func TestGetMetricHandler(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateGauge("testGauge", 123.45)

	r := setupRouter(storage)
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/testGauge", http.NoBody)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	expected := "123.45"
	if strings.TrimSpace(string(body)) != expected {
		t.Errorf("expected body to contain %q; got %q", expected, body)
	}
}

// TestInvalidRequest tests the response for an invalid route.
func TestInvalidRequest(t *testing.T) {
	storage := NewMemStorage()
	r := setupRouter(storage)
	req := httptest.NewRequest(http.MethodGet, "/invalid", http.NoBody)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %v; got %v", http.StatusNotFound, res.StatusCode)
	}
}
