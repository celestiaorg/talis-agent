package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// promMetrics is the singleton instance of PrometheusMetrics
	promMetrics *PrometheusMetrics
)

// PrometheusMetrics holds all Prometheus metrics for the agent
type PrometheusMetrics struct {
	// System metrics
	cpuUsage    prometheus.Gauge
	memoryUsage prometheus.Gauge
	diskUsage   prometheus.Gauge

	// Agent metrics
	uptime           prometheus.Counter
	checkinTimestamp prometheus.Gauge
	payloadReceived  prometheus.Counter
	commandSuccess   prometheus.Counter
	commandFailure   prometheus.Counter
}

// GetPrometheusMetrics returns the singleton instance of PrometheusMetrics
func GetPrometheusMetrics() *PrometheusMetrics {
	if promMetrics == nil {
		promMetrics = newPrometheusMetrics()
	}
	return promMetrics
}

// newPrometheusMetrics creates a new PrometheusMetrics instance
func newPrometheusMetrics() *PrometheusMetrics {
	pm := &PrometheusMetrics{
		// System metrics
		cpuUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		}),
		memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_memory_usage_percent",
			Help: "Current memory usage percentage",
		}),
		diskUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_disk_usage_percent",
			Help: "Current disk usage percentage",
		}),

		// Agent metrics
		uptime: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_uptime_seconds",
			Help: "Total uptime of the agent in seconds",
		}),
		checkinTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "agent_last_checkin_timestamp",
			Help: "Timestamp of the last successful check-in",
		}),
		payloadReceived: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_payload_bytes_received",
			Help: "Total number of bytes received in payloads",
		}),
		commandSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_command_executions_success",
			Help: "Number of successful command executions",
		}),
		commandFailure: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "agent_command_executions_failure",
			Help: "Number of failed command executions",
		}),
	}

	// Register all metrics
	prometheus.MustRegister(pm.cpuUsage)
	prometheus.MustRegister(pm.memoryUsage)
	prometheus.MustRegister(pm.diskUsage)
	prometheus.MustRegister(pm.uptime)
	prometheus.MustRegister(pm.checkinTimestamp)
	prometheus.MustRegister(pm.payloadReceived)
	prometheus.MustRegister(pm.commandSuccess)
	prometheus.MustRegister(pm.commandFailure)

	return pm
}

// UpdateSystemMetrics updates the system-related Prometheus metrics
func (pm *PrometheusMetrics) UpdateSystemMetrics(metrics *SystemMetrics) {
	pm.cpuUsage.Set(metrics.CPU.UsagePercent)
	pm.memoryUsage.Set(metrics.Memory.UsedPercent)
	pm.diskUsage.Set(metrics.Disk.UsedPercent)
}

// RecordUptime increments the uptime counter
func (pm *PrometheusMetrics) RecordUptime(seconds float64) {
	pm.uptime.Add(seconds)
}

// RecordCheckin updates the last check-in timestamp
func (pm *PrometheusMetrics) RecordCheckin(timestamp float64) {
	pm.checkinTimestamp.Set(timestamp)
}

// RecordPayloadReceived increments the payload bytes counter
func (pm *PrometheusMetrics) RecordPayloadReceived(bytes int64) {
	pm.payloadReceived.Add(float64(bytes))
}

// RecordCommandExecution records a command execution result
func (pm *PrometheusMetrics) RecordCommandExecution(success bool) {
	if success {
		pm.commandSuccess.Inc()
	} else {
		pm.commandFailure.Inc()
	}
}
