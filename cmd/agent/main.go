package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

var (
	flagAddr      string
	flagReportInt int
	flagPollInt   int
)

type Metrics struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func init() {
	defaultReportInt := 10
	defaultPollInt := 2

	flag.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server endpoint address (default: localhost:8080)")
	flag.IntVar(&flagReportInt, "r", defaultReportInt, "Report interval in seconds (default: 10)")
	flag.IntVar(&flagPollInt, "p", defaultPollInt, "Poll interval in seconds (default: 2)")

	if len(flag.Args()) > 0 {
		fmt.Println("Unknown flags:", flag.Args())
		flag.Usage()
		panic("Terminating due to unknown flags")
	}
}

func newMetrics() *Metrics {
	return &Metrics{
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

func (m *Metrics) updateMetrics() {
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

func (m *Metrics) sendMetric(metricType, name string, value interface{}) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", flagAddr, metricType, name, value)

	req, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Error closing response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected response code: %d\n", resp.StatusCode)
	}
}

func (m *Metrics) reportMetrics() {
	for name, value := range m.Gauges {
		m.sendMetric("gauge", name, strconv.FormatFloat(value, 'f', -1, 64))
	}
	for name, value := range m.Counters {
		m.sendMetric("counter", name, value)
	}
}

func main() {
	flag.Parse()

	metrics := newMetrics()

	tickerPoll := time.NewTicker(time.Duration(flagPollInt) * time.Second)
	tickerReport := time.NewTicker(time.Duration(flagReportInt) * time.Second)

	for {
		select {
		case <-tickerPoll.C:
			metrics.updateMetrics()
		case <-tickerReport.C:
			metrics.reportMetrics()
		}
	}
}
