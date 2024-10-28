package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// Определение переменных подключения.
var (
	host = "localhost"
	port = "8080"
)

// MetricType определение типов метрик.
type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

// Storage интерфейс для взаимодействия с хранилищем метрик.
type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAllMetrics() map[string]interface{}
}

// MemStorage cтруктура хранилища метрик.
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

// NewMemStorage cоздание нового хранилища.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge обновление метрик gauge и counter.
func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *MemStorage) UpdateCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

// GetAllMetrics получение всех метрик.
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

// updateMetricHandler основная функция обновления метрик.
func updateMetricHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")

		const expectedPartsLength = 5

		if len(parts) != expectedPartsLength {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}

		metricType := parts[2]
		metricName := parts[3]
		metricValue := parts[4]

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

// rootHandler обработчик для корневого URL.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, `<html>
    <body>
        <h2>Yandex practicum exporter</h2>
        <p><a href="/metrics">View Metrics</a></p>
    </body>
</html>`)
}

// getAllMetricsHandler обработчик отображения страницы метрик.
func getAllMetricsHandler(storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		metrics := storage.GetAllMetrics()

		jsonResponse, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(jsonResponse); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	address := fmt.Sprintf("%s:%s", host, port)

	storage := NewMemStorage()
	http.HandleFunc("/update/", updateMetricHandler(storage))
	http.HandleFunc("/metrics", getAllMetricsHandler(storage))
	http.HandleFunc("/", rootHandler)

	log.Printf("Server is running on http://%s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
