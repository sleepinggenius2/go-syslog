package transport

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"

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

func (t *BaseTransport) setMetrics(logParts message.LogParts, msgLen int64, err error) {
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

func (t *BaseTransport) LoadMetrics(metrics map[string]*dto.MetricFamily) {
	metricsMap := map[string]*prometheus.CounterVec{
		"errored":  t.metrics.MessagesErrored,
		"received": t.metrics.MessagesReceived,
	}
	for k, v := range metricsMap {
		totals := metrics["syslog_transport_messages_"+k+"_total"]
		if totals == nil || totals.GetType() != dto.MetricType_COUNTER {
			continue
		}
		for _, metric := range totals.GetMetric() {
			labels := metric.GetLabel()
			if len(labels) != 2 || labels[0] == nil || labels[1] == nil {
				continue
			}
			if labels[0].GetName() == "transport" && labels[0].GetValue() == t.network+":"+t.addr {
				if labels[1].GetName() == "client" {
					v.With(prometheus.Labels{"client": labels[1].GetValue()}).Add(metric.Counter.GetValue())
				}
			} else if labels[1].GetName() == "transport" && labels[1].GetValue() == t.network+":"+t.addr {
				if labels[0].GetName() == "client" {
					v.With(prometheus.Labels{"client": labels[0].GetValue()}).Add(metric.Counter.GetValue())
				}
			}
		}
	}
}
