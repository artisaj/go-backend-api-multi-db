package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsRecording(t *testing.T) {
	m := NewMetrics(10)

	m.RecordQuery(QueryMetric{
		DataSource: "test-ds",
		Table:      "users",
		Status:     "success",
		Latency:    50,
		Rows:       10,
	})

	m.RecordQuery(QueryMetric{
		DataSource: "test-ds",
		Table:      "users",
		Status:     "error",
		Latency:    20,
		Rows:       0,
	})

	metrics := m.GetMetrics()
	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, "test-ds", metrics[0].DataSource)
}

func TestMetricsSummary(t *testing.T) {
	m := NewMetrics(100)

	for i := 0; i < 10; i++ {
		m.RecordQuery(QueryMetric{
			DataSource: "ds1",
			Table:      "table1",
			Status:     "success",
			Latency:    int64(10 + i),
			Rows:       5,
		})
	}

	m.RecordQuery(QueryMetric{
		DataSource: "ds1",
		Table:      "table1",
		Status:     "error",
		Latency:    100,
		Rows:       0,
	})

	summary := m.GetSummary()
	assert.Equal(t, 1, len(summary))
	assert.Equal(t, "ds1", summary[0].DataSource)
	assert.Equal(t, 11, summary[0].TotalQueries)
	assert.Equal(t, 10, summary[0].SuccessCount)
	assert.Equal(t, 1, summary[0].ErrorCount)
	assert.Equal(t, int64(50), summary[0].TotalRows)
}

func TestMetricsMaxSize(t *testing.T) {
	m := NewMetrics(3)

	for i := 0; i < 5; i++ {
		m.RecordQuery(QueryMetric{
			DataSource: "test",
			Status:     "success",
			Latency:    10,
		})
		time.Sleep(1 * time.Millisecond)
	}

	metrics := m.GetMetrics()
	assert.Equal(t, 3, len(metrics), "Should keep only last 3 queries")
}
