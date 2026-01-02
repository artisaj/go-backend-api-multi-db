package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"

	"api-database/internal/domain/job"
	"api-database/internal/telemetry"
)

// QueryJobMessage é enviado via RabbitMQ.
type QueryJobMessage struct {
	ID          string       `json:"id"`
	PayloadHash string       `json:"payloadHash"`
	APIKey      string       `json:"apiKey"`
	DataSource  string       `json:"dataSource"`
	Table       string       `json:"table"`
	Request     QueryRequest `json:"request"`
	CreatedAt   time.Time    `json:"createdAt"`
}

// JobProcessor consome mensagens e executa queries.
type JobProcessor struct {
	service *QueryService
	jobs    job.JobRepository
	metrics MetricsRecorder
	logger  zerolog.Logger
}

// MetricsRecorder é interface mínima para registrar métricas sem acoplamento.
type MetricsRecorder interface {
	RecordQuery(metric telemetry.QueryMetric)
}

func NewJobProcessor(service *QueryService, jobs job.JobRepository, metrics MetricsRecorder, logger zerolog.Logger) *JobProcessor {
	return &JobProcessor{service: service, jobs: jobs, metrics: metrics, logger: logger}
}

// Handle decodifica a mensagem, executa a query e atualiza o job no repositório.
func (p *JobProcessor) Handle(ctx context.Context, body []byte) error {
	var msg QueryJobMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		p.logger.Error().Err(err).Msg("[WORKER] failed to unmarshal job message")
		return err
	}

	p.logger.Info().
		Str("job_id", msg.ID).
		Str("payload_hash", msg.PayloadHash).
		Str("data_source", msg.DataSource).
		Str("table", msg.Table).
		Str("api_key", msg.APIKey).
		Msg("[WORKER] processing job from queue")

	start := time.Now()
	_ = p.jobs.UpdateStatus(ctx, msg.ID, job.StatusRunning, map[string]any{
		"startedAt": time.Now(),
	})

	resp, err := p.service.QueryTable(ctx, msg.DataSource, msg.Table, msg.Request)
	tookMs := time.Since(start).Milliseconds()

	if err != nil {
		p.logger.Error().
			Err(err).
			Str("job_id", msg.ID).
			Int64("took_ms", tookMs).
			Msg("[WORKER] job failed")
		_ = p.jobs.UpdateStatus(ctx, msg.ID, job.StatusFailed, map[string]any{
			"error":      err.Error(),
			"finishedAt": time.Now(),
			"tookMs":     tookMs,
		})
		p.recordMetric(msg, "error", 0, tookMs)
		return err
	}

	p.logger.Info().
		Str("job_id", msg.ID).
		Int("rows", resp.Metadata.Rows).
		Int64("took_ms", resp.Metadata.TookMs).
		Msg("[WORKER] job completed successfully")

	_ = p.jobs.UpdateStatus(ctx, msg.ID, job.StatusSucceeded, map[string]any{
		"rows":       resp.Metadata.Rows,
		"tookMs":     resp.Metadata.TookMs,
		"finishedAt": time.Now(),
	})

	p.recordMetric(msg, "success", resp.Metadata.Rows, resp.Metadata.TookMs)
	return nil
}

func (p *JobProcessor) recordMetric(msg QueryJobMessage, status string, rows int, tookMs int64) {
	if p.metrics == nil {
		return
	}
	p.metrics.RecordQuery(telemetry.QueryMetric{
		DataSource:  msg.DataSource,
		Table:       msg.Table,
		Status:      status,
		Latency:     tookMs,
		Rows:        rows,
		JobID:       msg.ID,
		PayloadHash: msg.PayloadHash,
		APIKey:      msg.APIKey,
	})
}
