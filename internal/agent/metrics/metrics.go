package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	"github.com/avointsev/yp7m-go/internal/logger"
)

type MetricType struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

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
	m.Gauges["RandomValue"] = rand.Float64() * randomGaugeMultiplexor
	// runtime counter metrics
	m.Counters["PollCount"]++
}

func (m *MetricType) SendMetric(destAddress string, metricatype string, name string, value interface{}) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", destAddress, metricatype, name, value)

	fmt.Printf(url)
	
	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		log.Printf("%s: %v", logger.ErrAgentCreateRequest, err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s: %v", logger.ErrAgentSendRequest, err)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("%s: %v", logger.ErrAgentCloseRequest, closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("%s: %d", logger.ErrAgentResponseCode, resp.StatusCode)
	}
}

func (m *MetricType) ReportMetrics(destAddress string) {
	for name, value := range m.Gauges {
		m.SendMetric(destAddress, "gauge", name, strconv.FormatFloat(value, 'f', -1, 64))
	}
	for name, value := range m.Counters {
		m.SendMetric(destAddress, "counter", name, value)
	}
}
