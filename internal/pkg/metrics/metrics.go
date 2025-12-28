package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// MessagesProcessedTotal tracks total number of messages processed
	MessagesProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumer_messages_processed_total",
			Help: "Total number of messages processed",
		},
		[]string{"command", "status"},
	)

	// MessageProcessingDuration tracks message processing duration
	MessageProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_message_processing_duration_seconds",
			Help:    "Duration of message processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)

	// KafkaConsumerLag tracks consumer lag
	KafkaConsumerLag = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag",
			Help: "Current consumer lag",
		},
		[]string{"topic", "partition"},
	)

	// HTTPRequestsTotal tracks HTTP requests
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_consumer_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration tracks HTTP request duration
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_consumer_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"method", "endpoint"},
	)
)

// Handler returns the Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordMessageProcessed records a processed message
func RecordMessageProcessed(command, status string) {
	MessagesProcessedTotal.WithLabelValues(command, status).Inc()
}

// ObserveMessageProcessingDuration records message processing duration
func ObserveMessageProcessingDuration(command string, duration float64) {
	MessageProcessingDuration.WithLabelValues(command).Observe(duration)
}

// SetConsumerLag sets the consumer lag for a partition
func SetConsumerLag(topic string, partition int32, lag float64) {
	KafkaConsumerLag.WithLabelValues(topic, string(rune(partition))).Set(lag)
}
