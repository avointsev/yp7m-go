// Package handlers functions related with web handlers.
package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/avointsev/yp7m-go/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

const (
	// Gauge is the storage key for gauge metrics.
	Gauge = "gauge"
	// Counter is the storage key for counter metrics.
	Counter = "counter"
)

// UpdateMetricHandler function of update metrics.
func UpdateMetricHandler(store storage.Type) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" {
			http.Error(w, "Metric not found", http.StatusNotFound)
			log.Printf("%s: not pointed metric value", "Metric not found")
			return
		}

		var responseMessage string

		switch metricType {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				log.Printf("%s: %s", "Invalid gauge value", metricValue)
				return
			}
			store.UpdateGauge(metricName, value)
			responseMessage = "Metric " + metricName + " updated successfully"

		case Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				log.Printf("%s: %s", "Invalid counter value", metricValue)
				return
			}
			store.UpdateCounter(metricName, value)
			responseMessage = "Metric " + metricName + " updated successfully"

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			log.Printf("%s: %s", "Invalid metric type", metricType)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(responseMessage)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			log.Printf("%s for metric %s: %v", "Failed to write response", metricName, err)
			return
		}
		log.Println(responseMessage)
	}
}

// GetMetricHandler function getting metric.
func GetMetricHandler(store storage.Type) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		value, err := store.GetMetric(metricType, metricName)
		if err != nil {
			switch err.Error() {
			case "invalid metric type":
				http.Error(w, "Invalid metric type", http.StatusNotFound)
				return
			case "metric not found":
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			default:
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")

		switch metricType {
		case "gauge":
			floatValue, ok := value.(float64)
			if !ok {
				http.Error(w, "Invalid metric type", http.StatusInternalServerError)
				log.Printf("%s: expected float64 but got %T", "Invalid metric type", value)
				return
			}
			valueStr := strconv.FormatFloat(floatValue, 'f', -1, 64)
			_, err = w.Write([]byte(valueStr))
			if err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				log.Printf("%s for metric %s: %v", "Failed to write response", metricName, err)
				return
			}
		case "counter":
			intValue, ok := value.(int64)
			if !ok {
				http.Error(w, "Invalid metric type", http.StatusInternalServerError)
				log.Printf("%s: expected int64 but got %T", "Invalid metric type", value)
				return
			}
			valueStr := strconv.FormatInt(intValue, 10)
			_, err = w.Write([]byte(valueStr))
			if err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
				log.Printf("%s for metric %s: %v", "Failed to write response", metricName, err)
				return
			}
		default:
			http.Error(w, "Invalid metric type", http.StatusNotFound)
			return
		}
	}
}

// RootHandler main web handler.
func RootHandler(store storage.Type) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		tmpl, err := template.New("metrics").Parse(`
			<html>
				<body>
					<h2>Metrics List</h2>
					<ul>
						{{range $name, $value := .}}
							<li>{{$name}}: {{$value}}</li>
						{{end}}
					</ul>
				</body>
			</html>
		`)
		if err != nil {
			http.Error(w, "Failed to parse template", http.StatusInternalServerError)
			log.Printf("%s: %s", "Failed to parse template", err)
			return
		}

		metrics := store.GetAllMetrics()
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, metrics); err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			log.Printf("%s: %v", "Failed to execute template", err)
		}
	}
}
