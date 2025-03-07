package handlers

import (
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
)

// MetricsHandler handles metrics collection and reporting
type MetricsHandler struct {
	registry    *prometheus.Registry
	cpuUsage    prometheus.Gauge
	memoryUsage prometheus.Gauge
	goroutines  prometheus.Gauge

	// Disk I/O metrics
	diskReadBytes    prometheus.Gauge
	diskWriteBytes   prometheus.Gauge
	diskReadOps      prometheus.Gauge
	diskWriteOps     prometheus.Gauge
	diskReadLatency  prometheus.Gauge
	diskWriteLatency prometheus.Gauge

	// Network metrics
	networkBytesReceived prometheus.Gauge
	networkBytesSent     prometheus.Gauge
	networkPacketsRecv   prometheus.Gauge
	networkPacketsSent   prometheus.Gauge
	networkErrorsRecv    prometheus.Gauge
	networkErrorsSent    prometheus.Gauge
}

// NewMetricsHandler creates a new metrics handler instance
func NewMetricsHandler() *MetricsHandler {
	registry := prometheus.NewRegistry()

	cpuUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})

	memoryUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	goroutines := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_goroutines",
		Help: "Number of goroutines",
	})

	// Disk I/O metrics
	diskReadBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_read_bytes",
		Help: "Total bytes read from disk",
	})

	diskWriteBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_write_bytes",
		Help: "Total bytes written to disk",
	})

	diskReadOps := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_read_ops",
		Help: "Total number of disk read operations",
	})

	diskWriteOps := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_write_ops",
		Help: "Total number of disk write operations",
	})

	diskReadLatency := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_read_latency_seconds",
		Help: "Average disk read latency in seconds",
	})

	diskWriteLatency := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_write_latency_seconds",
		Help: "Average disk write latency in seconds",
	})

	// Network metrics
	networkBytesReceived := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_received",
		Help: "Total bytes received over network",
	})

	networkBytesSent := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_bytes_sent",
		Help: "Total bytes sent over network",
	})

	networkPacketsRecv := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_packets_received",
		Help: "Total packets received over network",
	})

	networkPacketsSent := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_packets_sent",
		Help: "Total packets sent over network",
	})

	networkErrorsRecv := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_errors_received",
		Help: "Total network errors received",
	})

	networkErrorsSent := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_network_errors_sent",
		Help: "Total network errors sent",
	})

	// Register all metrics
	registry.MustRegister(cpuUsage)
	registry.MustRegister(memoryUsage)
	registry.MustRegister(goroutines)
	registry.MustRegister(diskReadBytes)
	registry.MustRegister(diskWriteBytes)
	registry.MustRegister(diskReadOps)
	registry.MustRegister(diskWriteOps)
	registry.MustRegister(diskReadLatency)
	registry.MustRegister(diskWriteLatency)
	registry.MustRegister(networkBytesReceived)
	registry.MustRegister(networkBytesSent)
	registry.MustRegister(networkPacketsRecv)
	registry.MustRegister(networkPacketsSent)
	registry.MustRegister(networkErrorsRecv)
	registry.MustRegister(networkErrorsSent)

	return &MetricsHandler{
		registry:             registry,
		cpuUsage:             cpuUsage,
		memoryUsage:          memoryUsage,
		goroutines:           goroutines,
		diskReadBytes:        diskReadBytes,
		diskWriteBytes:       diskWriteBytes,
		diskReadOps:          diskReadOps,
		diskWriteOps:         diskWriteOps,
		diskReadLatency:      diskReadLatency,
		diskWriteLatency:     diskWriteLatency,
		networkBytesReceived: networkBytesReceived,
		networkBytesSent:     networkBytesSent,
		networkPacketsRecv:   networkPacketsRecv,
		networkPacketsSent:   networkPacketsSent,
		networkErrorsRecv:    networkErrorsRecv,
		networkErrorsSent:    networkErrorsSent,
	}
}

// Handle processes metrics requests
func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	// Update existing metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	h.goroutines.Set(float64(runtime.NumGoroutine()))
	h.memoryUsage.Set(float64(m.Alloc))

	// Update disk I/O metrics
	if diskStats, err := disk.IOCounters(); err == nil {
		var totalReadBytes, totalWriteBytes, totalReadOps, totalWriteOps float64
		var totalReadLatency, totalWriteLatency float64
		var count float64

		for _, stat := range diskStats {
			totalReadBytes += float64(stat.ReadBytes)
			totalWriteBytes += float64(stat.WriteBytes)
			totalReadOps += float64(stat.ReadCount)
			totalWriteOps += float64(stat.WriteCount)
			if stat.ReadTime > 0 {
				totalReadLatency += float64(stat.ReadTime) / float64(stat.ReadCount)
			}
			if stat.WriteTime > 0 {
				totalWriteLatency += float64(stat.WriteTime) / float64(stat.WriteCount)
			}
			count++
		}

		h.diskReadBytes.Set(totalReadBytes)
		h.diskWriteBytes.Set(totalWriteBytes)
		h.diskReadOps.Set(totalReadOps)
		h.diskWriteOps.Set(totalWriteOps)
		if count > 0 {
			h.diskReadLatency.Set(totalReadLatency / count)
			h.diskWriteLatency.Set(totalWriteLatency / count)
		}
	}

	// Update network metrics
	if netStats, err := net.IOCounters(false); err == nil {
		var totalBytesRecv, totalBytesSent, totalPacketsRecv, totalPacketsSent float64
		var totalErrorsRecv, totalErrorsSent float64

		for _, stat := range netStats {
			totalBytesRecv += float64(stat.BytesRecv)
			totalBytesSent += float64(stat.BytesSent)
			totalPacketsRecv += float64(stat.PacketsRecv)
			totalPacketsSent += float64(stat.PacketsSent)
			totalErrorsRecv += float64(stat.Errin)
			totalErrorsSent += float64(stat.Errout)
		}

		h.networkBytesReceived.Set(totalBytesRecv)
		h.networkBytesSent.Set(totalBytesSent)
		h.networkPacketsRecv.Set(totalPacketsRecv)
		h.networkPacketsSent.Set(totalPacketsSent)
		h.networkErrorsRecv.Set(totalErrorsRecv)
		h.networkErrorsSent.Set(totalErrorsSent)
	}

	// Return metrics in Prometheus format
	metrics, err := h.registry.Gather()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to gather metrics",
		})
	}

	return c.JSON(fiber.Map{
		"metrics": metrics,
	})
}
