package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRootHandler Test root handler.
func TestRootHandler(t *testing.T) {
	rec := httptest.NewRecorder()

	rootHandler(rec)

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

// TestGetAllMetricsHandler Test getting of all metrics.
func TestGetAllMetricsHandler(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateGauge("testGauge", 123.45)
	storage.UpdateCounter("testCounter", 100)

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	rec := httptest.NewRecorder()

	getAllMetricsHandler(storage)(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %v; got %v", http.StatusOK, res.StatusCode)
	}

	var metrics map[string]interface{}
	err := json.NewDecoder(res.Body).Decode(&metrics)
	if err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if metrics["testGauge"] != 123.45 || metrics["testCounter"] != float64(100) {
		t.Errorf("expected metrics to contain testGauge=123.45 and testCounter=100; got %v", metrics)
	}
}

// TestUpdateMetricHandler Test metic update.
func TestUpdateMetricHandler(t *testing.T) {
	storage := NewMemStorage()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/123.45", http.NoBody)
	rec := httptest.NewRecorder()

	updateMetricHandler(storage)(rec, req)

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

	// Checking that metica has been updated
	if storage.GetAllMetrics()["testGauge"] != 123.45 {
		t.Errorf("expected testGauge to be 123.45; got %v", storage.GetAllMetrics()["testGauge"])
	}
}

// TestInvalidMethodHandler Test uncorrect request.
func TestInvalidMethodHandler(t *testing.T) {
	storage := NewMemStorage()
	req := httptest.NewRequest(http.MethodGet, "/update/gauge/testGauge/123.45", http.NoBody)
	rec := httptest.NewRecorder()

	updateMetricHandler(storage)(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status %v; got %v", http.StatusMethodNotAllowed, res.StatusCode)
	}
}

// TestDefaultHandlerInvalidPath Test unknown handler paths.
func TestDefaultHandlerInvalidPath(t *testing.T) {
	storage := NewMemStorage()
	req := httptest.NewRequest(http.MethodGet, "/unknown", http.NoBody)
	rec := httptest.NewRecorder()

	defaultHandler(storage)(rec, req)

	res := rec.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Errorf("could not close response body: %v", err)
		}
	}()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %v; got %v", http.StatusBadRequest, res.StatusCode)
	}
}
