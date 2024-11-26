package main

import (
	_ "bytes"
	_ "log"
	"net/http"
	"net/http/httptest"
	_ "os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/avointsev/yp7m-go/internal/server/handlers"
	"github.com/avointsev/yp7m-go/internal/server/storage"
)

func TestMainFunction(t *testing.T) {
	mockStorage := storage.NewMemStorage()

	// var logBuffer bytes.Buffer
	// log.SetOutput(&logBuffer)
	// defer log.SetOutput(os.Stderr)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", handlers.RootHandler(mockStorage))
	r.Get("/value/{type}/{name}", handlers.GetMetricHandler(mockStorage))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(mockStorage))

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %v", resp.StatusCode)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	resp, err = http.Get(ts.URL + "/value/gauge/test_metric")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", resp.StatusCode)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/gauge/test_metric/10.5", http.NoBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %v", resp.StatusCode)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("error closing response body: %v", closeErr)
		}
	}()

	// logOutput := logBuffer.String()
	// if logOutput != "" {
	// 	t.Errorf("unexpected log output: %v", logOutput)
	// }
}
