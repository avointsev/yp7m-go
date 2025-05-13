// Package metrics function releted with metrics
package metrics

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
)

const (
	// Number of bits in a float64 mantissa (precision of math/rand.Float64).
	randMantissaBits = 53
	// Bits to shift right on a 64-bit value to get randMantissaBits of entropy.
	randShiftBits = 64 - randMantissaBits
)

// MetricType define metric structure.
type MetricType struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// NewMetrics set init metrics state.
func NewMetrics() *MetricType {
	return &MetricType{
		Gauges: map[string]float64{
			"Alloc":         0,
			"BuckHashSys":   0,
			"Frees":         0,
			"GCCPUFraction": 0,
			"GCSys":         0,
			"HeapAlloc":     0,
			"HeapIdle":      0,
			"HeapInuse":     0,
			"HeapObjects":   0,
			"HeapReleased":  0,
			"HeapSys":       0,
			"LastGC":        0,
			"Lookups":       0,
			"MCacheInuse":   0,
			"MCacheSys":     0,
			"MSpanInuse":    0,
			"MSpanSys":      0,
			"Mallocs":       0,
			"NextGC":        0,
			"NumForcedGC":   0,
			"NumGC":         0,
			"OtherSys":      0,
			"PauseTotalNs":  0,
			"StackInuse":    0,
			"StackSys":      0,
			"Sys":           0,
			"TotalAlloc":    0,
			"RandomValue":   0,
		},
		Counters: map[string]int64{
			"PollCount": 0,
		},
	}
}

// cryptoFloat64 generate rundom number.
func cryptoFloat64() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Printf("crypto/rand error: %v", err)
		return 0
	}
	u := binary.LittleEndian.Uint64(b[:]) >> randShiftBits
	return float64(u) / (1 << randMantissaBits)
}

// UpdateMetrics update metrics.
func (m *MetricType) UpdateMetrics() {
	var stats runtime.MemStats
	const randomGaugeMultiplexor = 100.0
	runtime.ReadMemStats(&stats)

	// runtime gauge metrics
	m.Gauges["Alloc"] = float64(stats.Alloc)
	m.Gauges["BuckHashSys"] = float64(stats.BuckHashSys)
	m.Gauges["Frees"] = float64(stats.Frees)
	m.Gauges["GCCPUFraction"] = stats.GCCPUFraction
	m.Gauges["GCSys"] = float64(stats.GCSys)
	m.Gauges["HeapAlloc"] = float64(stats.HeapAlloc)
	m.Gauges["HeapIdle"] = float64(stats.HeapIdle)
	m.Gauges["HeapInuse"] = float64(stats.HeapInuse)
	m.Gauges["HeapObjects"] = float64(stats.HeapObjects)
	m.Gauges["HeapReleased"] = float64(stats.HeapReleased)
	m.Gauges["HeapSys"] = float64(stats.HeapSys)
	m.Gauges["LastGC"] = float64(stats.LastGC)
	m.Gauges["Lookups"] = float64(stats.Lookups)
	m.Gauges["MCacheInuse"] = float64(stats.MCacheInuse)
	m.Gauges["MCacheSys"] = float64(stats.MCacheSys)
	m.Gauges["MSpanInuse"] = float64(stats.MSpanInuse)
	m.Gauges["MSpanSys"] = float64(stats.MSpanSys)
	m.Gauges["Mallocs"] = float64(stats.Mallocs)
	m.Gauges["NextGC"] = float64(stats.NextGC)
	m.Gauges["NumForcedGC"] = float64(stats.NumForcedGC)
	m.Gauges["NumGC"] = float64(stats.NumGC)
	m.Gauges["OtherSys"] = float64(stats.OtherSys)
	m.Gauges["PauseTotalNs"] = float64(stats.PauseTotalNs)
	m.Gauges["StackInuse"] = float64(stats.StackInuse)
	m.Gauges["StackSys"] = float64(stats.StackSys)
	m.Gauges["Sys"] = float64(stats.Sys)
	m.Gauges["TotalAlloc"] = float64(stats.TotalAlloc)
	// random gauge metrics
	m.Gauges["RandomValue"] = cryptoFloat64() * randomGaugeMultiplexor
	// runtime counter metrics
	m.Counters["PollCount"]++
}

// SendMetric function for send metric to server.
func (m *MetricType) SendMetric(destAddress string, metricatype string, name string, value interface{}) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", destAddress, metricatype, name, value)

	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Printf("%s: %v", "Error creating request", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s: %v", "Error sending request", err)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("%s: %v", "Error closing response body", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("%s: %d", "Unexpected response code", resp.StatusCode)
	}
}

// ReportMetrics push metrics.
func (m *MetricType) ReportMetrics(destAddress string) {
	for name, value := range m.Gauges {
		m.SendMetric(destAddress, "gauge", name, strconv.FormatFloat(value, 'f', -1, 64))
	}
	for name, value := range m.Counters {
		m.SendMetric(destAddress, "counter", name, value)
	}
}
