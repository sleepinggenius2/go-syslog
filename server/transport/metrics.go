package transport

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sleepinggenius2/go-syslog/common/message"
)

type Metrics struct {
	MessageLength    prometheus.Histogram
	MessagesErrored  *prometheus.CounterVec
	MessagesReceived *prometheus.CounterVec
}

func (t *BaseTransport) addMetrics() {
	t.metrics = &Metrics{
		MessagesErrored: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "syslog",
				Subsystem:   "transport",
				Name:        "messages_errored_total",
				Help:        "Total number of errored messages per client.",
				ConstLabels: prometheus.Labels{"transport": t.network + ":" + t.addr},
			},
			[]string{"client"},
		),
		MessageLength: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace:   "syslog",
				Subsystem:   "transport",
				Name:        "message_length",
				Help:        "Total number of messages received per client.",
				ConstLabels: prometheus.Labels{"transport": t.network + ":" + t.addr},
				Buckets:     prometheus.ExponentialBuckets(64, 2, 11),
			},
		),
		MessagesReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   "syslog",
				Subsystem:   "transport",
				Name:        "messages_received_total",
				Help:        "Total number of messages received per client.",
				ConstLabels: prometheus.Labels{"transport": t.network + ":" + t.addr},
			},
			[]string{"client"},
		),
	}
}

func (t BaseTransport) setMetrics(logParts message.LogParts, msgLen int64, err error) {
	if t.metrics == nil {
		return
	}

	t.metrics.MessageLength.Observe(float64(msgLen))

	labels := prometheus.Labels{"client": logParts.Client.Host}
	t.metrics.MessagesReceived.With(labels).Inc()
	if err != nil {
		t.metrics.MessagesErrored.With(labels).Inc()
	}
}
