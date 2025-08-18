package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ProducerMetrics contains all Prometheus metrics for the producer
type ProducerMetrics struct {
	MessagesProduced    prometheus.CounterVec
	MessagesErrors      prometheus.CounterVec
	ProductionLatency   prometheus.HistogramVec
	EventProcessingTime prometheus.HistogramVec
	ActiveGoroutines    prometheus.Gauge
	QueueSize           prometheus.Gauge
}

// NewProducerMetrics creates and registers Prometheus metrics
func NewProducerMetrics() *ProducerMetrics {
	return &ProducerMetrics{
		MessagesProduced: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_producer_messages_produced_total",
				Help: "Total number of messages successfully produced to Kafka",
			},
			[]string{"topic", "partition"},
		),
		MessagesErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_producer_messages_errors_total",
				Help: "Total number of message production errors",
			},
			[]string{"topic", "error_type", "error_code"},
		),
		ProductionLatency: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_producer_production_latency_seconds",
				Help:    "Time taken to produce a message to Kafka",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"topic"},
		),
		EventProcessingTime: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kafka_producer_event_processing_time_seconds",
				Help:    "Time taken to process success/error events",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"event_type"},
		),
		ActiveGoroutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kafka_producer_active_goroutines",
				Help: "Number of active goroutines in the producer",
			},
		),
		QueueSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "kafka_producer_queue_size",
				Help: "Current size of the producer input queue",
			},
		),
	}
}

// RecordMessageProduced increments the success counter
func (m *ProducerMetrics) RecordMessageProduced(topic string, partition int32) {
	m.MessagesProduced.WithLabelValues(topic, fmt.Sprintf("%d", partition)).Inc()
}

// RecordMessageError increments the error counter
func (m *ProducerMetrics) RecordMessageError(topic, errorType, errorCode string) {
	m.MessagesErrors.WithLabelValues(topic, errorType, errorCode).Inc()
}

// RecordProductionLatency records the time taken to produce a message
func (m *ProducerMetrics) RecordProductionLatency(topic string, latency float64) {
	m.ProductionLatency.WithLabelValues(topic).Observe(latency)
}

// RecordEventProcessingTime records the time taken to process an event
func (m *ProducerMetrics) RecordEventProcessingTime(eventType string, duration float64) {
	m.EventProcessingTime.WithLabelValues(eventType).Observe(duration)
}

// SetActiveGoroutines sets the number of active goroutines
func (m *ProducerMetrics) SetActiveGoroutines(count float64) {
	m.ActiveGoroutines.Set(count)
}

// SetQueueSize sets the current queue size
func (m *ProducerMetrics) SetQueueSize(size float64) {
	m.QueueSize.Set(size)
}
