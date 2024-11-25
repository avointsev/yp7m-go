package storage

import (
	"testing"
)

func TestUpdateGauge(t *testing.T) {
	memStorage := NewMemStorage()

	memStorage.UpdateGauge("gauge_metric", 10.5)
	value, err := memStorage.GetMetric("gauge", "gauge_metric")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 10.5 {
		t.Errorf("expected value 10.5, got %v", value)
	}
}

func TestUpdateCounter(t *testing.T) {
	memStorage := NewMemStorage()

	memStorage.UpdateCounter("counter_metric", 5)
	value, err := memStorage.GetMetric("counter", "counter_metric")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != int64(6) {
		t.Errorf("expected value 6, got %v", value)
	}
}

func TestGetAllMetrics(t *testing.T) {
	memStorage := NewMemStorage()

	memStorage.UpdateGauge("gauge_metric", 10.5)
	memStorage.UpdateCounter("counter_metric", 5)

	allMetrics := memStorage.GetAllMetrics()

	if allMetrics["gauge_metric"] != 10.5 {
		t.Errorf("expected value 10.5 for gauge_metric, got %v", allMetrics["gauge_metric"])
	}
	if allMetrics["counter_metric"] != int64(6) {
		t.Errorf("expected value 6 for counter_metric, got %v", allMetrics["counter_metric"])
	}
}

func TestGetMetricNotFound(t *testing.T) {
	memStorage := NewMemStorage()

	_, err := memStorage.GetMetric("gauge", "non_existent_metric")
	if err == nil {
		t.Errorf("expected an error for non-existent metric, got nil")
	}
}

func TestGetMetricInvalidType(t *testing.T) {
	memStorage := NewMemStorage()

	memStorage.UpdateGauge("gauge_metric", 10.5)
	_, err := memStorage.GetMetric("counter", "gauge_metric")
	if err == nil {
		t.Errorf("expected an error for invalid metric type, got nil")
	}
}
