package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/avointsev/yp7m-go/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

// setupRouter creates and returns a new chi router for testing.
func setupRouter(store storage.StorageType) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", RootHandler(store))
	r.Get("/value/{type}/{name}", GetMetricHandler(store))
	r.Post("/update/{type}/{name}/{value}", UpdateMetricHandler(store))
	return r
}

// TestRootHandler tests the root handler.
func TestRootHandler(t *testing.T) {
	store := storage.NewMemStorage()
	store.UpdateGauge("testGauge", 123.45)
	store.UpdateCounter("testCounter", 100)

	r := setupRouter(store)
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
	store := storage.NewMemStorage()
	r := setupRouter(store)
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

	if store.GetAllMetrics()["testGauge"] != 123.45 {
		t.Errorf("expected testGauge to be 123.45; got %v", store.GetAllMetrics()["testGauge"])
	}
}

// TestGetMetricHandler tests retrieving a metric value.
func TestGetMetricHandler(t *testing.T) {
	store := storage.NewMemStorage()
	store.UpdateGauge("testGauge", 123.45)

	r := setupRouter(store)
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
	store := storage.NewMemStorage()
	r := setupRouter(store)
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
