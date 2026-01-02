package httpserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"api-database/internal/application/data"
	"api-database/internal/domain"
	"api-database/internal/domain/job"
	"api-database/internal/infrastructure/rabbitmq"
	httpmiddleware "api-database/internal/presentation/http/middleware"
	"api-database/internal/telemetry"
)

// DataHandler expõe endpoints de dados.
type DataHandler struct {
	service *data.QueryService
	metrics *telemetry.Metrics
	jobs    job.JobRepository
	queue   *rabbitmq.Client
}

func NewDataHandler(service *data.QueryService, metrics *telemetry.Metrics, jobs job.JobRepository, queue *rabbitmq.Client) *DataHandler {
	return &DataHandler{service: service, metrics: metrics, jobs: jobs, queue: queue}
}

func (h *DataHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	table := chi.URLParam(r, "table")

	var req data.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		appErr := domain.NewAppError(domain.ErrInvalidInput, "invalid request body", http.StatusBadRequest)
		writeError(w, appErr)
		return
	}

	asyncRequested := strings.EqualFold(r.URL.Query().Get("async"), "true") || strings.EqualFold(r.Header.Get("Prefer"), "respond-async")
	if asyncRequested && h.queue != nil && h.jobs != nil {
		h.enqueueAsync(w, r, source, table, req)
		return
	}

	resp, err := h.service.QueryTable(r.Context(), source, table, req)
	if err != nil {
		appErr, ok := err.(*domain.AppError)
		if !ok {
			appErr = domain.NewAppError(domain.ErrInternal, err.Error(), http.StatusInternalServerError)
		}
		writeError(w, appErr)

		// Registrar métrica de erro
		if h.metrics != nil {
			h.metrics.RecordQuery(telemetry.QueryMetric{
				DataSource:  source,
				Table:       table,
				Status:      "error",
				Latency:     0,
				Rows:        0,
				JobID:       "",
				PayloadHash: "",
				APIKey:      h.apiKey(r),
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)

	// Registrar métrica de sucesso
	if h.metrics != nil {
		h.metrics.RecordQuery(telemetry.QueryMetric{
			DataSource:  source,
			Table:       table,
			Status:      "success",
			Latency:     resp.Metadata.TookMs,
			Rows:        resp.Metadata.Rows,
			JobID:       "",
			PayloadHash: "",
			APIKey:      h.apiKey(r),
		})
	}
}

// enqueueAsync cria job, publica na fila e retorna 202.
func (h *DataHandler) enqueueAsync(w http.ResponseWriter, r *http.Request, source, table string, req data.QueryRequest) {
	if h.queue == nil || h.jobs == nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "async queue unavailable", http.StatusServiceUnavailable))
		return
	}

	payloadHash, err := data.HashQueryRequest(req)
	if err != nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "failed to hash request", http.StatusInternalServerError))
		return
	}

	jobID := uuid.NewString()
	now := time.Now()
	apiKey := h.apiKey(r)
	jobRecord := &job.QueryJob{
		ID:          jobID,
		PayloadHash: payloadHash,
		APIKey:      apiKey,
		DataSource:  source,
		Table:       table,
		Status:      job.StatusQueued,
		CreatedAt:   now,
	}

	if err := h.jobs.Insert(r.Context(), jobRecord); err != nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "failed to persist job", http.StatusInternalServerError))
		return
	}

	msg := data.QueryJobMessage{
		ID:          jobID,
		PayloadHash: payloadHash,
		APIKey:      apiKey,
		DataSource:  source,
		Table:       table,
		Request:     req,
		CreatedAt:   now,
	}

	if err := h.queue.Publish(r.Context(), msg); err != nil {
		_ = h.jobs.UpdateStatus(r.Context(), jobID, job.StatusFailed, map[string]any{"error": "queue publish failed", "finishedAt": time.Now()})
		writeError(w, domain.NewAppError(domain.ErrInternal, "failed to enqueue job", http.StatusInternalServerError))
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"jobId":       jobID,
		"status":      job.StatusQueued,
		"payloadHash": payloadHash,
		"message":     "Job enqueued for async processing",
	})
}

// HandleJobStatus retorna o estado de um job específico.
func (h *DataHandler) HandleJobStatus(w http.ResponseWriter, r *http.Request) {
	if h.jobs == nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "job repository unavailable", http.StatusServiceUnavailable))
		return
	}
	jobID := chi.URLParam(r, "jobID")
	record, err := h.jobs.GetByID(r.Context(), jobID)
	if err != nil || record == nil {
		writeError(w, domain.NewAppError(domain.ErrNotFound, "job not found", http.StatusNotFound))
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// HandleJobsByHash retorna jobs associados a um hash de payload.
func (h *DataHandler) HandleJobsByHash(w http.ResponseWriter, r *http.Request) {
	if h.jobs == nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "job repository unavailable", http.StatusServiceUnavailable))
		return
	}
	hash := chi.URLParam(r, "hash")
	items, err := h.jobs.GetByPayloadHash(r.Context(), hash)
	if err != nil {
		writeError(w, domain.NewAppError(domain.ErrInternal, "failed to fetch jobs", http.StatusInternalServerError))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *DataHandler) apiKey(r *http.Request) string {
	ak := httpmiddleware.GetAPIKeyFromContext(r.Context())
	if ak == nil {
		return ""
	}
	return ak.Key
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, appErr *domain.AppError) {
	status := appErr.Status()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(appErr)
}
