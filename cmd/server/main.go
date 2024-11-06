package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	flagAddr string
)

// MetricType defines metric types.
type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

// Storage interface for interacting with MemStorage.
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAllMetrics() map[string]interface{}
	GetMetric(metricType, name string) (interface{}, bool)
}

// MemStorage memory storage for metrics.
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

func init() {
	flag.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server address (default: localhost:8080)")

	if len(flag.Args()) > 0 {
		fmt.Println("Unknown flags:", flag.Args())
		flag.Usage()
		panic("Terminating due to unknown flags")
	}
}

// NewMemStorage creates a new instance of MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge updates the value of a gauge metric.
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// UpdateCounter updates the value of a counter metric.
func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

// GetAllMetrics returns a map of all available metrics.
func (m *MemStorage) GetAllMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	allMetrics := make(map[string]interface{})
	for name, value := range m.gauges {
		allMetrics[name] = value
	}
	for name, value := range m.counters {
		allMetrics[name] = value
	}
	return allMetrics
}

// GetMetric returns the value of a specific metric by type and name.
func (m *MemStorage) GetMetric(metricType, name string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch MetricType(metricType) {
	case Gauge:
		value, exists := m.gauges[name]
		return value, exists
	case Counter:
		value, exists := m.counters[name]
		return value, exists
	default:
		return nil, false
	}
}

// updateMetricHandler handles updating metrics.
func updateMetricHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		var responseMessage string

		switch MetricType(metricType) {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)
			responseMessage = fmt.Sprintf("Metric %s updated successfully", metricName)

		case Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)
			responseMessage = fmt.Sprintf("Metric %s updated successfully", metricName)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(responseMessage)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

// getMetricHandler handles retrieving the value of a specific metric.
func getMetricHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		value, exists := storage.GetMetric(metricType, metricName)
		if !exists {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%v", value)
	}
}

// rootHandler handles the root URL and returns a list of all metrics.
func rootHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		metrics := storage.GetAllMetrics()
		fmt.Fprintln(w, "<html><body><h2>Metrics List</h2><ul>")
		for name, value := range metrics {
			fmt.Fprintf(w, "<li>%s: %v</li>", name, value)
		}
		fmt.Fprintln(w, "</ul></body></html>")
	}
}

func main() {
	flag.Parse()

	storage := NewMemStorage()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", rootHandler(storage))
	r.Get("/value/{type}/{name}", getMetricHandler(storage))
	r.Post("/update/{type}/{name}/{value}", updateMetricHandler(storage))

	log.Printf("Server is running on http://%s", flagAddr)
	if err := http.ListenAndServe(flagAddr, r); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
