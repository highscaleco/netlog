package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// NetworkBytesTotal is a counter for the total number of bytes transferred
	NetworkBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "netlog_network_bytes_total",
			Help: "Total number of bytes transferred",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port", "direction"},
	)

	// NetworkPacketsTotal is a counter for the total number of packets
	NetworkPacketsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "netlog_network_packets_total",
			Help: "Total number of packets",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port", "direction"},
	)

	// NetworkConnectionsActive is a gauge for the number of active connections
	NetworkConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "netlog_network_connections_active",
			Help: "Number of active connections",
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port"},
	)

	// NetworkConnectionDuration is a histogram for the duration of network connections
	NetworkConnectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "netlog_network_connection_duration_seconds",
			Help:    "Duration of network connections in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"namespace", "name", "source", "destination", "protocol", "port"},
	)

	// Track active metrics for cleanup
	activeMetrics     = make(map[string]time.Time)
	activeMetricsLock sync.RWMutex
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
	labels := prometheus.Labels{
		"namespace":   namespace,
		"name":        name,
		"source":      source,
		"destination": destination,
		"protocol":    protocol,
		"port":        port,
		"direction":   direction,
	}

	// Update counters
	NetworkBytesTotal.With(labels).Add(float64(bytes))
	NetworkPacketsTotal.With(labels).Add(float64(packets))

	// Update connection metrics
	connLabels := prometheus.Labels{
		"namespace":   namespace,
		"name":        name,
		"source":      source,
		"destination": destination,
		"protocol":    protocol,
		"port":        port,
	}

	NetworkConnectionsActive.With(connLabels).Inc()
	NetworkConnectionDuration.With(connLabels).Observe(duration)

	// Track metric for cleanup
	metricKey := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s", namespace, name, source, destination, protocol, port, direction)
	activeMetricsLock.Lock()
	activeMetrics[metricKey] = time.Now()
	activeMetricsLock.Unlock()
}

// CleanupMetrics removes metrics that haven't been updated recently
func CleanupMetrics() {
	activeMetricsLock.Lock()
	defer activeMetricsLock.Unlock()

	now := time.Now()
	for key, lastUpdate := range activeMetrics {
		if now.Sub(lastUpdate) > 5*time.Minute {
			// Parse the key to get labels
			parts := strings.Split(key, ":")
			if len(parts) == 7 {
				labels := prometheus.Labels{
					"namespace":   parts[0],
					"name":        parts[1],
					"source":      parts[2],
					"destination": parts[3],
					"protocol":    parts[4],
					"port":        parts[5],
					"direction":   parts[6],
				}

				// Remove metrics
				NetworkBytesTotal.Delete(labels)
				NetworkPacketsTotal.Delete(labels)

				connLabels := prometheus.Labels{
					"namespace":   parts[0],
					"name":        parts[1],
					"source":      parts[2],
					"destination": parts[3],
					"protocol":    parts[4],
					"port":        parts[5],
				}
				NetworkConnectionsActive.Delete(connLabels)
				NetworkConnectionDuration.Delete(connLabels)

				// Remove from active metrics
				delete(activeMetrics, key)
			}
		}
	}
}
