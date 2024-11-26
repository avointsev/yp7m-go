package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/avointsev/yp7m-go/internal/logger"
	"github.com/avointsev/yp7m-go/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

func UpdateMetricHandler(store storage.StorageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" {
			http.Error(w, logger.ErrMetricNotFound, http.StatusNotFound)
			log.Printf("%s: not pointed metric value", logger.ErrMetricNotFound)
			return
		}

		var responseMessage string

		switch metricType {
		case Gauge:
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, logger.ErrMetricInvalidGaugeValue, http.StatusNotFound)
				log.Printf(logger.LogDefaultFormat, logger.ErrMetricInvalidGaugeValue, metricValue)
				return
			}
			store.UpdateGauge(metricName, value)
			responseMessage = "Metric " + metricName + " " + logger.OkUpdated

		case Counter:
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, logger.ErrMetricInvalidCounterValue, http.StatusNotFound)
				log.Printf(logger.LogDefaultFormat, logger.ErrMetricInvalidCounterValue, metricValue)
				return
			}
			if value < 0 {
				http.Error(w, logger.ErrMetricInvalidCounterValue, http.StatusNotFound)
				log.Printf("%s: counter value cannot be negative, %s == %d", logger.ErrMetricInvalidCounterValue, metricName, value)
				return
			}
			store.UpdateCounter(metricName, value)
			responseMessage = "Metric " + metricName + " " + logger.OkUpdated

		default:
			http.Error(w, logger.ErrMetricInvalidType, http.StatusNotFound)
			log.Printf(logger.LogDefaultFormat, logger.ErrMetricInvalidType, metricType)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(responseMessage)); err != nil {
			http.Error(w, logger.ErrWriteResponce, http.StatusInternalServerError)
			log.Printf("%s for metric %s: %v", logger.ErrWriteResponce, metricName, err)
			return
		}
		log.Println(responseMessage)
	}
}

func GetMetricHandler(store storage.StorageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		value, err := store.GetMetric(metricType, metricName)
		if err != nil {
			switch err.Error() {
			case logger.ErrMetricInvalidType:
				http.Error(w, logger.ErrMetricInvalidType, http.StatusNotFound)
				return
			case logger.ErrMetricNotFound:
				http.Error(w, logger.ErrMetricNotFound, http.StatusNotFound)
				return
			default:
				http.Error(w, logger.ErrServerInternalError, http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain")

		switch metricType {
		case "gauge":
			floatValue, ok := value.(float64)
			if !ok {
				http.Error(w, logger.ErrMetricInvalidType, http.StatusInternalServerError)
				log.Printf("%s: expected float64 but got %T", logger.ErrMetricInvalidType, value)
				return
			}
			valueStr := strconv.FormatFloat(floatValue, 'f', -1, 64)
			_, err = w.Write([]byte(valueStr))
			if err != nil {
				http.Error(w, logger.ErrWriteResponce, http.StatusInternalServerError)
				log.Printf("%s for metric %s: %v", logger.ErrWriteResponce, metricName, err)
				return
			}
		case "counter":
			intValue, ok := value.(int64)
			if !ok {
				http.Error(w, logger.ErrMetricInvalidType, http.StatusInternalServerError)
				log.Printf("%s: expected int64 but got %T", logger.ErrMetricInvalidType, value)
				return
			}
			valueStr := strconv.FormatInt(intValue, 10)
			_, err = w.Write([]byte(valueStr))
			if err != nil {
				http.Error(w, logger.ErrWriteResponce, http.StatusInternalServerError)
				log.Printf("%s for metric %s: %v", logger.ErrWriteResponce, metricName, err)
				return
			}
		default:
			http.Error(w, logger.ErrMetricInvalidType, http.StatusNotFound)
			return
		}
	}
}

func RootHandler(store storage.StorageType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, logger.ErrHTMLTemplateParse, http.StatusInternalServerError)
			log.Printf("%s: %s", logger.ErrHTMLTemplateParse, err)
			return
		}

		metrics := store.GetAllMetrics()
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, metrics); err != nil {
			http.Error(w, logger.ErrHTMLTemplateExecute, http.StatusInternalServerError)
			log.Printf("%s: %v", logger.ErrHTMLTemplateExecute, err)
		}
	}
}
