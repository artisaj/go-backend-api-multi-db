package job

import (
	"context"
	"time"
)

// Status indica o estado de processamento do job.
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusFailed    = "failed"
	StatusSucceeded = "succeeded"
)

// QueryJob representa uma solicitação de consulta para processamento assíncrono.
type QueryJob struct {
	ID          string     `bson:"_id" json:"id"`
	PayloadHash string     `bson:"payloadHash" json:"payloadHash"`
	APIKey      string     `bson:"apiKey" json:"apiKey"`
	DataSource  string     `bson:"dataSource" json:"dataSource"`
	Table       string     `bson:"table" json:"table"`
	Status      string     `bson:"status" json:"status"`
	Rows        int        `bson:"rows" json:"rows"`
	TookMs      int64      `bson:"tookMs" json:"tookMs"`
	Error       string     `bson:"error" json:"error"`
	CreatedAt   time.Time  `bson:"createdAt" json:"createdAt"`
	StartedAt   *time.Time `bson:"startedAt" json:"startedAt,omitempty"`
	FinishedAt  *time.Time `bson:"finishedAt" json:"finishedAt,omitempty"`
}

// JobRepository define operações para persistir jobs.
type JobRepository interface {
	Insert(ctx context.Context, job *QueryJob) error
	UpdateStatus(ctx context.Context, id string, status string, fields map[string]any) error
	GetByID(ctx context.Context, id string) (*QueryJob, error)
	GetByPayloadHash(ctx context.Context, payloadHash string) ([]*QueryJob, error)
}
