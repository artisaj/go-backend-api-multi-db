package telemetry

import (
	"sync"
	"time"
)

// QueryMetric rastreia uma query executada.
type QueryMetric struct {
	DataSource string
	Table      string
	Status     string // "success", "error"
	Latency    int64  // ms
	Rows       int
	Timestamp  time.Time
}

// Metrics coleta métricas de queries.
type Metrics struct {
	mu      sync.RWMutex
	queries []QueryMetric
	maxSize int // manter últimas N queries
}

// NewMetrics cria um novo collector de métricas.
func NewMetrics(maxSize int) *Metrics {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &Metrics{
		queries: make([]QueryMetric, 0, maxSize),
		maxSize: maxSize,
	}
}

// RecordQuery registra uma query executada.
func (m *Metrics) RecordQuery(metric QueryMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metric.Timestamp = time.Now()
	m.queries = append(m.queries, metric)

	// Manter apenas últimas N queries
	if len(m.queries) > m.maxSize {
		m.queries = m.queries[len(m.queries)-m.maxSize:]
	}
}

// GetMetrics retorna snapshot das métricas.
func (m *Metrics) GetMetrics() []QueryMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]QueryMetric, len(m.queries))
	copy(result, m.queries)
	return result
}

// Summary retorna resumo de métricas por datasource.
type Summary struct {
	DataSource   string `json:"dataSource"`
	TotalQueries int    `json:"totalQueries"`
	SuccessCount int    `json:"successCount"`
	ErrorCount   int    `json:"errorCount"`
	AvgLatencyMs int64  `json:"avgLatencyMs"`
	P95LatencyMs int64  `json:"p95LatencyMs"`
	TotalRows    int64  `json:"totalRows"`
}

// GetSummary retorna resumo agregado por datasource.
func (m *Metrics) GetSummary() []Summary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make(map[string]*Summary)

	for _, q := range m.queries {
		if _, ok := summaries[q.DataSource]; !ok {
			summaries[q.DataSource] = &Summary{
				DataSource: q.DataSource,
			}
		}

		s := summaries[q.DataSource]
		s.TotalQueries++
		if q.Status == "success" {
			s.SuccessCount++
		} else {
			s.ErrorCount++
		}
		s.TotalRows += int64(q.Rows)
	}

	// Calcular latências
	for ds, s := range summaries {
		var latencies []int64
		for _, q := range m.queries {
			if q.DataSource == ds {
				latencies = append(latencies, q.Latency)
			}
		}

		if len(latencies) > 0 {
			sum := int64(0)
			for _, lat := range latencies {
				sum += lat
			}
			s.AvgLatencyMs = sum / int64(len(latencies))

			// P95
			if len(latencies) > 1 {
				// Quick sort seria melhor, mas para simplicidade...
				for i := 0; i < len(latencies); i++ {
					for j := i + 1; j < len(latencies); j++ {
						if latencies[j] < latencies[i] {
							latencies[i], latencies[j] = latencies[j], latencies[i]
						}
					}
				}
				idx := (len(latencies) * 95) / 100
				if idx < len(latencies) {
					s.P95LatencyMs = latencies[idx]
				}
			}
		}
	}

	result := make([]Summary, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, *s)
	}
	return result
}
