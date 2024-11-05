package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupRouter(storage Storage) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", rootHandler(storage))
	r.Get("/value/{type}/{name}", getMetricHandler(storage))
	r.Post("/update/{type}/{name}/{value}", updateMetricHandler(storage))
	return r
}

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

	if storage.GetAllMetrics()["testGauge"] != 123.45 {
		t.Errorf("expected testGauge to be 123.45; got %v", storage.GetAllMetrics()["testGauge"])
	}
}

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
