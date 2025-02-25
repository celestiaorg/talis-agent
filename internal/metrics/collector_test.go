package metrics

import (
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	interval := 5 * time.Second
	collector := NewCollector(interval)

	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}

	if collector.interval != interval {
		t.Errorf("Expected interval %v, got %v", interval, collector.interval)
	}
}

func TestCollect(t *testing.T) {
	collector := NewCollector(time.Second)
	metrics, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Verify that we got non-nil metrics
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	// Verify timestamp is recent
	if time.Since(metrics.Timestamp) > time.Minute {
		t.Error("Timestamp is too old")
	}

	// Basic validation of CPU metrics
	if metrics.CPU.UsagePercent < 0 || metrics.CPU.UsagePercent > 100 {
		t.Errorf("Invalid CPU usage percentage: %v", metrics.CPU.UsagePercent)
	}

	// Basic validation of memory metrics
	if metrics.Memory.Total == 0 {
		t.Error("Expected non-zero total memory")
	}
	if metrics.Memory.UsedPercent < 0 || metrics.Memory.UsedPercent > 100 {
		t.Errorf("Invalid memory usage percentage: %v", metrics.Memory.UsedPercent)
	}

	// Basic validation of disk metrics
	if metrics.Disk.Total == 0 {
		t.Error("Expected non-zero total disk space")
	}
	if metrics.Disk.UsedPercent < 0 || metrics.Disk.UsedPercent > 100 {
		t.Errorf("Invalid disk usage percentage: %v", metrics.Disk.UsedPercent)
	}

	// Basic validation of network metrics
	if len(metrics.Network.Interfaces) == 0 {
		t.Error("Expected at least one network interface")
	}

	// Basic validation of host info
	if metrics.HostInfo.Hostname == "" {
		t.Error("Expected non-empty hostname")
	}
	if metrics.HostInfo.OS == "" {
		t.Error("Expected non-empty OS")
	}
	if metrics.HostInfo.Platform == "" {
		t.Error("Expected non-empty platform")
	}
}
