package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMetrics represents the collected system metrics
type SystemMetrics struct {
	Timestamp time.Time     `json:"timestamp"`
	CPU       CPUMetrics    `json:"cpu"`
	Memory    MemoryMetrics `json:"memory"`
	Disk      DiskMetrics   `json:"disk"`
	Network   NetMetrics    `json:"network"`
	HostInfo  HostInfo      `json:"host_info"`
}

// CPUMetrics represents CPU-related metrics
type CPUMetrics struct {
	UsagePercent float64   `json:"usage_percent"`
	PerCPU       []float64 `json:"per_cpu"`
}

// MemoryMetrics represents memory-related metrics
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskMetrics represents disk-related metrics
type DiskMetrics struct {
	Total       uint64                         `json:"total"`
	Used        uint64                         `json:"used"`
	Free        uint64                         `json:"free"`
	UsedPercent float64                        `json:"used_percent"`
	IOCounters  map[string]disk.IOCountersStat `json:"io_counters"`
}

// NetMetrics represents network-related metrics
type NetMetrics struct {
	Interfaces []string                      `json:"interfaces"`
	IOCounters map[string]net.IOCountersStat `json:"io_counters"`
}

// HostInfo represents host-related information
type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Uptime   uint64 `json:"uptime"`
}

// Collector implements prometheus.Collector interface
type Collector struct {
	interval time.Duration

	// CPU metrics
	cpuUsage   *prometheus.Desc
	cpuPerCore *prometheus.Desc

	// Memory metrics
	memoryTotal   *prometheus.Desc
	memoryUsed    *prometheus.Desc
	memoryFree    *prometheus.Desc
	memoryPercent *prometheus.Desc

	// Disk metrics
	diskTotal   *prometheus.Desc
	diskUsed    *prometheus.Desc
	diskFree    *prometheus.Desc
	diskPercent *prometheus.Desc
	diskIO      *prometheus.Desc

	// Network metrics
	networkIO *prometheus.Desc

	// Host metrics
	hostUptime *prometheus.Desc
}

// NewCollector creates a new metrics collector
func NewCollector(interval time.Duration) *Collector {
	return &Collector{
		interval: interval,

		// CPU metrics
		cpuUsage: prometheus.NewDesc(
			"system_cpu_usage_percent",
			"Current CPU usage percentage",
			nil, nil,
		),
		cpuPerCore: prometheus.NewDesc(
			"system_cpu_core_usage_percent",
			"CPU usage percentage per core",
			[]string{"core"}, nil,
		),

		// Memory metrics
		memoryTotal: prometheus.NewDesc(
			"system_memory_total_bytes",
			"Total memory in bytes",
			nil, nil,
		),
		memoryUsed: prometheus.NewDesc(
			"system_memory_used_bytes",
			"Used memory in bytes",
			nil, nil,
		),
		memoryFree: prometheus.NewDesc(
			"system_memory_free_bytes",
			"Free memory in bytes",
			nil, nil,
		),
		memoryPercent: prometheus.NewDesc(
			"system_memory_usage_percent",
			"Memory usage percentage",
			nil, nil,
		),

		// Disk metrics
		diskTotal: prometheus.NewDesc(
			"system_disk_total_bytes",
			"Total disk space in bytes",
			nil, nil,
		),
		diskUsed: prometheus.NewDesc(
			"system_disk_used_bytes",
			"Used disk space in bytes",
			nil, nil,
		),
		diskFree: prometheus.NewDesc(
			"system_disk_free_bytes",
			"Free disk space in bytes",
			nil, nil,
		),
		diskPercent: prometheus.NewDesc(
			"system_disk_usage_percent",
			"Disk usage percentage",
			nil, nil,
		),
		diskIO: prometheus.NewDesc(
			"system_disk_io_bytes",
			"Disk I/O in bytes",
			[]string{"device", "type"}, nil,
		),

		// Network metrics
		networkIO: prometheus.NewDesc(
			"system_network_io_bytes",
			"Network I/O in bytes",
			[]string{"interface", "direction"}, nil,
		),

		// Host metrics
		hostUptime: prometheus.NewDesc(
			"system_uptime_seconds",
			"System uptime in seconds",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsage
	ch <- c.cpuPerCore
	ch <- c.memoryTotal
	ch <- c.memoryUsed
	ch <- c.memoryFree
	ch <- c.memoryPercent
	ch <- c.diskTotal
	ch <- c.diskUsed
	ch <- c.diskFree
	ch <- c.diskPercent
	ch <- c.diskIO
	ch <- c.networkIO
	ch <- c.hostUptime
}

// Collect implements prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	// Collect CPU metrics
	if percent, err := cpu.Percent(0, false); err == nil && len(percent) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.cpuUsage,
			prometheus.GaugeValue,
			percent[0],
		)
	}

	if perCPU, err := cpu.Percent(0, true); err == nil {
		for i, usage := range perCPU {
			ch <- prometheus.MustNewConstMetric(
				c.cpuPerCore,
				prometheus.GaugeValue,
				usage,
				fmt.Sprintf("%d", i),
			)
		}
	}

	// Collect memory metrics
	if v, err := mem.VirtualMemory(); err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.memoryTotal,
			prometheus.GaugeValue,
			float64(v.Total),
		)
		ch <- prometheus.MustNewConstMetric(
			c.memoryUsed,
			prometheus.GaugeValue,
			float64(v.Used),
		)
		ch <- prometheus.MustNewConstMetric(
			c.memoryFree,
			prometheus.GaugeValue,
			float64(v.Free),
		)
		ch <- prometheus.MustNewConstMetric(
			c.memoryPercent,
			prometheus.GaugeValue,
			v.UsedPercent,
		)
	}

	// Collect disk metrics
	if partitions, err := disk.Partitions(false); err == nil {
		for _, partition := range partitions {
			if usage, err := disk.Usage(partition.Mountpoint); err == nil {
				ch <- prometheus.MustNewConstMetric(
					c.diskTotal,
					prometheus.GaugeValue,
					float64(usage.Total),
				)
				ch <- prometheus.MustNewConstMetric(
					c.diskUsed,
					prometheus.GaugeValue,
					float64(usage.Used),
				)
				ch <- prometheus.MustNewConstMetric(
					c.diskFree,
					prometheus.GaugeValue,
					float64(usage.Free),
				)
				ch <- prometheus.MustNewConstMetric(
					c.diskPercent,
					prometheus.GaugeValue,
					usage.UsedPercent,
				)
				break // Only use root partition
			}
		}
	}

	// Collect disk I/O metrics
	if iostats, err := disk.IOCounters(); err == nil {
		for device, stats := range iostats {
			ch <- prometheus.MustNewConstMetric(
				c.diskIO,
				prometheus.GaugeValue,
				float64(stats.ReadBytes),
				device, "read",
			)
			ch <- prometheus.MustNewConstMetric(
				c.diskIO,
				prometheus.GaugeValue,
				float64(stats.WriteBytes),
				device, "write",
			)
		}
	}

	// Collect network metrics
	if netStats, err := net.IOCounters(true); err == nil {
		for _, stats := range netStats {
			ch <- prometheus.MustNewConstMetric(
				c.networkIO,
				prometheus.GaugeValue,
				float64(stats.BytesRecv),
				stats.Name, "received",
			)
			ch <- prometheus.MustNewConstMetric(
				c.networkIO,
				prometheus.GaugeValue,
				float64(stats.BytesSent),
				stats.Name, "sent",
			)
		}
	}

	// Collect host metrics
	if hostInfo, err := host.Info(); err == nil {
		ch <- prometheus.MustNewConstMetric(
			c.hostUptime,
			prometheus.GaugeValue,
			float64(hostInfo.Uptime),
		)
	}
}
