package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NetworkBytesTotal represents the total number of bytes transferred
	NetworkBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "netlog_network_bytes_total",
			Help: "Total number of bytes transferred",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port", "direction"},
	)

	// NetworkPacketsTotal represents the total number of packets
	NetworkPacketsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "netlog_network_packets_total",
			Help: "Total number of packets",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port", "direction"},
	)

	// NetworkConnectionsActive represents the number of active connections
	NetworkConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "netlog_network_connections_active",
			Help: "Number of active connections",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port"},
	)

	// NetworkConnectionDuration represents the duration of connections
	NetworkConnectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "netlog_network_connection_duration_seconds",
			Help:    "Duration of network connections in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port"},
	)
)

// Init initializes all metrics
func Init() {
	prometheus.MustRegister(NetworkBytesTotal)
	prometheus.MustRegister(NetworkPacketsTotal)
	prometheus.MustRegister(NetworkConnectionsActive)
	prometheus.MustRegister(NetworkConnectionDuration)
}

// UpdateMetrics updates all metrics based on the aggregated info
func UpdateMetrics(namespace, name, source, destination, protocol, port, direction string, bytes, packets int64, duration float64) {
	// Update counters
	NetworkBytesTotal.WithLabelValues(namespace, name, source, destination, protocol, port, direction).Add(float64(bytes))
	NetworkPacketsTotal.WithLabelValues(namespace, name, source, destination, protocol, port, direction).Add(float64(packets))

	// Update connection duration histogram
	NetworkConnectionDuration.WithLabelValues(namespace, name, source, destination, protocol, port).Observe(duration)

	// Update active connections gauge
	NetworkConnectionsActive.WithLabelValues(namespace, name, source, destination, protocol, port).Set(1)
}
