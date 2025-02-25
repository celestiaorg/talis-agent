package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestGetPrometheusMetrics(t *testing.T) {
	// Test singleton pattern
	metrics1 := GetPrometheusMetrics()
	metrics2 := GetPrometheusMetrics()

	if metrics1 != metrics2 {
		t.Error("GetPrometheusMetrics should return the same instance")
	}
}

func TestUpdateSystemMetrics(t *testing.T) {
	metrics := GetPrometheusMetrics()

	systemMetrics := &SystemMetrics{
		CPU: CPUMetrics{
			UsagePercent: 50.5,
		},
		Memory: MemoryMetrics{
			UsedPercent: 75.2,
		},
		Disk: DiskMetrics{
			UsedPercent: 85.7,
		},
	}

	metrics.UpdateSystemMetrics(systemMetrics)

	// Test CPU usage metric
	if value := testutil.ToFloat64(metrics.cpuUsage); value != 50.5 {
		t.Errorf("Expected CPU usage 50.5, got %v", value)
	}

	// Test memory usage metric
	if value := testutil.ToFloat64(metrics.memoryUsage); value != 75.2 {
		t.Errorf("Expected memory usage 75.2, got %v", value)
	}

	// Test disk usage metric
	if value := testutil.ToFloat64(metrics.diskUsage); value != 85.7 {
		t.Errorf("Expected disk usage 85.7, got %v", value)
	}
}

func TestRecordCommandExecution(t *testing.T) {
	metrics := GetPrometheusMetrics()

	// Test successful command
	initialTotal := testutil.ToFloat64(metrics.commandsTotal)
	initialSucceeded := testutil.ToFloat64(metrics.commandsSucceeded)

	metrics.RecordCommandExecution(true)

	if value := testutil.ToFloat64(metrics.commandsTotal); value != initialTotal+1 {
		t.Errorf("Expected commands total %v, got %v", initialTotal+1, value)
	}
	if value := testutil.ToFloat64(metrics.commandsSucceeded); value != initialSucceeded+1 {
		t.Errorf("Expected commands succeeded %v, got %v", initialSucceeded+1, value)
	}

	// Test failed command
	initialFailed := testutil.ToFloat64(metrics.commandsFailed)
	metrics.RecordCommandExecution(false)

	if value := testutil.ToFloat64(metrics.commandsTotal); value != initialTotal+2 {
		t.Errorf("Expected commands total %v, got %v", initialTotal+2, value)
	}
	if value := testutil.ToFloat64(metrics.commandsFailed); value != initialFailed+1 {
		t.Errorf("Expected commands failed %v, got %v", initialFailed+1, value)
	}
}

func TestRecordPayloadReceived(t *testing.T) {
	metrics := GetPrometheusMetrics()

	initialPayloads := testutil.ToFloat64(metrics.payloadsReceived)
	initialBytes := testutil.ToFloat64(metrics.payloadBytes)

	// Test payload recording
	metrics.RecordPayloadReceived(1024)

	if value := testutil.ToFloat64(metrics.payloadsReceived); value != initialPayloads+1 {
		t.Errorf("Expected payloads received %v, got %v", initialPayloads+1, value)
	}
	if value := testutil.ToFloat64(metrics.payloadBytes); value != initialBytes+1024 {
		t.Errorf("Expected payload bytes %v, got %v", initialBytes+1024, value)
	}
}

func TestRecordCheckin(t *testing.T) {
	metrics := GetPrometheusMetrics()

	timestamp := float64(time.Now().Unix())
	metrics.RecordCheckin(timestamp)

	if value := testutil.ToFloat64(metrics.lastCheckin); value != timestamp {
		t.Errorf("Expected last checkin timestamp %v, got %v", timestamp, value)
	}
}

func TestRecordUptime(t *testing.T) {
	metrics := GetPrometheusMetrics()

	initialUptime := testutil.ToFloat64(metrics.uptime)

	metrics.RecordUptime(60)
	if value := testutil.ToFloat64(metrics.uptime); value != initialUptime+60 {
		t.Errorf("Expected uptime %v, got %v", initialUptime+60, value)
	}

	metrics.RecordUptime(30)
	if value := testutil.ToFloat64(metrics.uptime); value != initialUptime+90 {
		t.Errorf("Expected uptime %v, got %v", initialUptime+90, value)
	}
}
