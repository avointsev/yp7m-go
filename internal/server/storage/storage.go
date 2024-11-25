package storage

import (
	"errors"
	"sync"

	"github.com/avointsev/yp7m-go/internal/logger"
)

// MetricType defines metric types.
type MetricType string

const (
	Gauge   = "gauge"
	Counter = "counter"
)

// StorageType interface for interacting with MemStorage.
type StorageType interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetAllMetrics() map[string]interface{}
	GetMetric(metricType, name string) (interface{}, error)
}

// MemStorage memory storage for metrics.
type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
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
	m.counters[name] = value + 1
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

func (m *MemStorage) GetMetric(metricType, name string) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch MetricType(metricType) {
	case Gauge:
		value, ok := m.gauges[name]
		if !ok {
			return nil, errors.New(logger.ErrMetricInvalidType)
		}
		return value, nil
	case Counter:
		value, ok := m.counters[name]
		if !ok {
			return nil, errors.New(logger.ErrMetricInvalidType)
		}
		return value, nil
	default:
		return nil, errors.New(logger.ErrMetricNotFound)
	}
}
