package main

import (
	"log"
	"time"

	"github.com/avointsev/yp7m-go/internal/agent/metrics"
	"github.com/avointsev/yp7m-go/internal/flags"
	"github.com/avointsev/yp7m-go/internal/logger"
)

func main() {
	config, err := flags.ParseAgentConfig()
	if err != nil {
		log.Fatalf("%s: %v", logger.ErrFlagsParse, err)
	}

	metricaSet := metrics.NewMetrics()

	tickerPoll := time.NewTicker(config.ReportInterval)
	tickerReport := time.NewTicker(config.PollInterval)

	for {
		select {
		case <-tickerPoll.C:
			metricaSet.UpdateMetrics()
		case <-tickerReport.C:
			metricaSet.ReportMetrics(config.Address)
		}
	}
}
