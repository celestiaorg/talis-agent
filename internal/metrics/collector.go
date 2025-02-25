package metrics

import (
	"fmt"
	"time"

	"github.com/celestiaorg/talis-agent/internal/logging"
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

// Collector handles system metrics collection
type Collector struct {
	interval time.Duration
}

// NewCollector creates a new metrics collector
func NewCollector(interval time.Duration) *Collector {
	logging.Debug().
		Str("interval", interval.String()).
		Msg("Creating new metrics collector")
	return &Collector{
		interval: interval,
	}
}

// Collect gathers all system metrics
func (c *Collector) Collect() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	var err error

	// Collect CPU metrics
	if err = c.collectCPU(&metrics.CPU); err != nil {
		logging.Error().Err(err).Msg("Failed to collect CPU metrics")
		return nil, fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// Collect memory metrics
	if err = c.collectMemory(&metrics.Memory); err != nil {
		logging.Error().Err(err).Msg("Failed to collect memory metrics")
		return nil, fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// Collect disk metrics
	if err = c.collectDisk(&metrics.Disk); err != nil {
		logging.Error().Err(err).Msg("Failed to collect disk metrics")
		return nil, fmt.Errorf("failed to collect disk metrics: %w", err)
	}

	// Collect network metrics
	if err = c.collectNetwork(&metrics.Network); err != nil {
		logging.Error().Err(err).Msg("Failed to collect network metrics")
		return nil, fmt.Errorf("failed to collect network metrics: %w", err)
	}

	// Collect host info
	if err = c.collectHostInfo(&metrics.HostInfo); err != nil {
		logging.Error().Err(err).Msg("Failed to collect host info")
		return nil, fmt.Errorf("failed to collect host info: %w", err)
	}

	logging.Debug().
		Float64("cpu_usage", metrics.CPU.UsagePercent).
		Float64("memory_usage", metrics.Memory.UsedPercent).
		Float64("disk_usage", metrics.Disk.UsedPercent).
		Int("network_interfaces", len(metrics.Network.Interfaces)).
		Msg("System metrics collected")

	return metrics, nil
}

func (c *Collector) collectCPU(metrics *CPUMetrics) error {
	percent, err := cpu.Percent(c.interval, false)
	if err != nil {
		return err
	}
	if len(percent) > 0 {
		metrics.UsagePercent = percent[0]
	}

	perCPU, err := cpu.Percent(c.interval, true)
	if err != nil {
		return err
	}
	metrics.PerCPU = perCPU

	logging.Debug().
		Float64("total_usage", metrics.UsagePercent).
		Int("cpu_count", len(metrics.PerCPU)).
		Msg("CPU metrics collected")

	return nil
}

func (c *Collector) collectMemory(metrics *MemoryMetrics) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	metrics.Total = v.Total
	metrics.Used = v.Used
	metrics.Free = v.Free
	metrics.UsedPercent = v.UsedPercent

	logging.Debug().
		Uint64("total", metrics.Total).
		Uint64("used", metrics.Used).
		Uint64("free", metrics.Free).
		Float64("used_percent", metrics.UsedPercent).
		Msg("Memory metrics collected")

	return nil
}

func (c *Collector) collectDisk(metrics *DiskMetrics) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	// We'll use the root partition for overall disk metrics
	for _, partition := range partitions {
		if partition.Mountpoint == "/" {
			usage, err := disk.Usage(partition.Mountpoint)
			if err != nil {
				return err
			}
			metrics.Total = usage.Total
			metrics.Used = usage.Used
			metrics.Free = usage.Free
			metrics.UsedPercent = usage.UsedPercent
			break
		}
	}

	// Collect IO counters
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return err
	}
	metrics.IOCounters = ioCounters

	logging.Debug().
		Uint64("total", metrics.Total).
		Uint64("used", metrics.Used).
		Uint64("free", metrics.Free).
		Float64("used_percent", metrics.UsedPercent).
		Int("io_counters", len(metrics.IOCounters)).
		Msg("Disk metrics collected")

	return nil
}

func (c *Collector) collectNetwork(metrics *NetMetrics) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	metrics.Interfaces = make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		metrics.Interfaces = append(metrics.Interfaces, iface.Name)
	}

	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	metrics.IOCounters = make(map[string]net.IOCountersStat)
	for _, counter := range ioCounters {
		metrics.IOCounters[counter.Name] = counter
	}

	logging.Debug().
		Int("interfaces", len(metrics.Interfaces)).
		Int("io_counters", len(metrics.IOCounters)).
		Msg("Network metrics collected")

	return nil
}

func (c *Collector) collectHostInfo(info *HostInfo) error {
	hostInfo, err := host.Info()
	if err != nil {
		return err
	}

	info.Hostname = hostInfo.Hostname
	info.OS = hostInfo.OS
	info.Platform = hostInfo.Platform
	info.Uptime = hostInfo.Uptime

	logging.Debug().
		Str("hostname", info.Hostname).
		Str("os", info.OS).
		Str("platform", info.Platform).
		Uint64("uptime", info.Uptime).
		Msg("Host info collected")

	return nil
}
