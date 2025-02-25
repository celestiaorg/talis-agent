package metrics

import (
	"runtime"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Singleton instance of PrometheusMetrics
	promMetrics     *PrometheusMetrics
	promMetricsOnce sync.Once
)

// PrometheusMetrics holds all Prometheus metrics
type PrometheusMetrics struct {
	// System metrics
	cpuUsage    prometheus.Gauge
	memoryUsage prometheus.Gauge
	diskUsage   prometheus.Gauge

	// Command execution metrics
	commandsTotal     prometheus.Counter
	commandsSucceeded prometheus.Counter
	commandsFailed    prometheus.Counter

	// Payload metrics
	payloadsReceived prometheus.Counter
	payloadBytes     prometheus.Counter

	// Agent status metrics
	lastCheckin prometheus.Gauge
	uptime      prometheus.Counter
}

// GetPrometheusMetrics returns the singleton instance of PrometheusMetrics
func GetPrometheusMetrics() *PrometheusMetrics {
	promMetricsOnce.Do(func() {
		promMetrics = newPrometheusMetrics()
	})
	return promMetrics
}

// newPrometheusMetrics creates and registers all Prometheus metrics
func newPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		// System metrics
		cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "talis_agent_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		}),
		memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "talis_agent_memory_usage_percent",
			Help: "Current memory usage percentage",
		}),
		diskUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "talis_agent_disk_usage_percent",
			Help: "Current disk usage percentage",
		}),

		// Command execution metrics
		commandsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_commands_total",
			Help: "Total number of commands executed",
		}),
		commandsSucceeded: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_commands_succeeded",
			Help: "Number of commands executed successfully",
		}),
		commandsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_commands_failed",
			Help: "Number of commands that failed execution",
		}),

		// Payload metrics
		payloadsReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_payloads_received_total",
			Help: "Total number of payloads received",
		}),
		payloadBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_payload_bytes_total",
			Help: "Total number of payload bytes received",
		}),

		// Agent status metrics
		lastCheckin: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "talis_agent_last_checkin_timestamp",
			Help: "Timestamp of the last successful check-in",
		}),
		uptime: promauto.NewCounter(prometheus.CounterOpts{
			Name: "talis_agent_uptime_seconds",
			Help: "Total uptime of the agent in seconds",
		}),
	}

	// Register version info metric
	promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "talis_agent_info",
			Help: "Information about the Talis agent",
		},
		[]string{"version", "os", "arch"},
	).With(prometheus.Labels{
		"version": "0.1.0", // TODO: Make this configurable
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}).Set(1)

	return metrics
}

// UpdateSystemMetrics updates the system metrics with the latest values
func (p *PrometheusMetrics) UpdateSystemMetrics(metrics *SystemMetrics) {
	p.cpuUsage.Set(metrics.CPU.UsagePercent)
	p.memoryUsage.Set(metrics.Memory.UsedPercent)
	p.diskUsage.Set(metrics.Disk.UsedPercent)
}

// RecordCommandExecution records a command execution attempt
func (p *PrometheusMetrics) RecordCommandExecution(succeeded bool) {
	p.commandsTotal.Inc()
	if succeeded {
		p.commandsSucceeded.Inc()
	} else {
		p.commandsFailed.Inc()
	}
}

// RecordPayloadReceived records a received payload
func (p *PrometheusMetrics) RecordPayloadReceived(bytes int64) {
	p.payloadsReceived.Inc()
	p.payloadBytes.Add(float64(bytes))
}

// RecordCheckin records a successful check-in
func (p *PrometheusMetrics) RecordCheckin(timestamp float64) {
	p.lastCheckin.Set(timestamp)
}

// RecordUptime increments the uptime counter
func (p *PrometheusMetrics) RecordUptime(seconds float64) {
	p.uptime.Add(seconds)
}
